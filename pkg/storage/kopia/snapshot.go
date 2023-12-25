package kopia

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/kopia/kopia/fs"
	"github.com/kopia/kopia/fs/virtualfs"
	"github.com/kopia/kopia/repo"
	"github.com/kopia/kopia/snapshot"
	"github.com/kopia/kopia/snapshot/policy"
	"github.com/kopia/kopia/snapshot/snapshotfs"
)

func snapshotSingleFile(ctx context.Context, fileName string, r io.Reader, rep repo.RepositoryWriter, tags map[string]string) (*snapshot.Manifest, error) {
	sourceInfo := snapshot.SourceInfo{
		Path:     "-",
		Host:     rep.ClientOptions().Hostname,
		UserName: rep.ClientOptions().Username,
	}
	fsEntry := virtualfs.NewStaticDirectory("-", []fs.Entry{
		virtualfs.StreamingFileFromReader(fileName, io.NopCloser(r)),
	})
	setManual := true

	log(ctx).Infof("Snapshotting %v ...", sourceInfo)

	previous, err := findPreviousSnapshotManifest(ctx, rep, sourceInfo, nil)
	if err != nil {
		return nil, err
	}

	policyTree, err := policy.TreeForSource(ctx, rep, sourceInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to get policy tree, %w", err)
	}

	u := setupUploader(ctx, rep)
	manifest, err := u.Upload(ctx, fsEntry, policyTree, sourceInfo, previous...)
	if err != nil {
		// fail-fast uploads will fail here without recording a manifest, other uploads will
		// possibly fail later.
		return nil, fmt.Errorf("upload error, %w", err)
	}

	manifest.Tags = tags

	if _, err = snapshot.SaveSnapshot(ctx, rep, manifest); err != nil {
		return nil, fmt.Errorf("cannot save manifest, %w", err)
	}

	if _, err = policy.ApplyRetentionPolicy(ctx, rep, sourceInfo, true); err != nil {
		return nil, fmt.Errorf("unable to apply retention policy, %w", err)
	}

	if setManual {
		if err = policy.SetManual(ctx, rep, sourceInfo); err != nil {
			return nil, fmt.Errorf("unable to set manual field in scheduling policy for source, %w", err)
		}
	}

	if ferr := rep.Flush(ctx); ferr != nil {
		return nil, fmt.Errorf("flush error, %w", ferr)
	}

	return manifest, reportSnapshotStatus(ctx, manifest)
}

func reportSnapshotStatus(ctx context.Context, manifest *snapshot.Manifest) error {
	var maybePartial string
	if manifest.IncompleteReason != "" {
		maybePartial = " partial"
	}

	sourceInfo := manifest.Source

	snapID := manifest.ID

	log(ctx).Infof("Created%v snapshot with root %v and ID %v in %v", maybePartial, manifest.RootObjectID(), snapID, manifest.EndTime.Sub(manifest.StartTime).Truncate(time.Second))

	if ds := manifest.RootEntry.DirSummary; ds != nil {
		if ds.IgnoredErrorCount > 0 {
			log(ctx).Warnf("Ignored %v error(s) while snapshotting %v.", ds.IgnoredErrorCount, sourceInfo)
		}

		if ds.FatalErrorCount > 0 {
			return fmt.Errorf("found %v fatal error(s) while snapshotting %v", ds.FatalErrorCount, sourceInfo) //nolint:revive
		}
	}

	return nil
}

// findPreviousSnapshotManifest returns the list of previous snapshots for a given source, including
// last complete snapshot and possibly some number of incomplete snapshots following it.
func findPreviousSnapshotManifest(ctx context.Context, rep repo.Repository, sourceInfo snapshot.SourceInfo, noLaterThan *fs.UTCTimestamp) ([]*snapshot.Manifest, error) {
	man, err := snapshot.ListSnapshots(ctx, rep, sourceInfo)
	if err != nil {
		return nil, fmt.Errorf("error listing previous snapshots, %w", err)
	}

	// phase 1 - find latest complete snapshot.
	var previousComplete *snapshot.Manifest

	var previousCompleteStartTime fs.UTCTimestamp

	var result []*snapshot.Manifest

	for _, p := range man {
		if noLaterThan != nil && p.StartTime.After(*noLaterThan) {
			continue
		}

		if p.IncompleteReason == "" && (previousComplete == nil || p.StartTime.After(previousComplete.StartTime)) {
			previousComplete = p
			previousCompleteStartTime = p.StartTime
		}
	}

	if previousComplete != nil {
		result = append(result, previousComplete)
	}

	// add all incomplete snapshots after that
	for _, p := range man {
		if noLaterThan != nil && p.StartTime.After(*noLaterThan) {
			continue
		}

		if p.IncompleteReason != "" && p.StartTime.After(previousCompleteStartTime) {
			result = append(result, p)
		}
	}

	return result, nil
}

func setupUploader(ctx context.Context, rep repo.RepositoryWriter) *snapshotfs.Uploader {
	u := snapshotfs.NewUploader(rep)

	onCtrlC(u.Cancel)

	u.ForceHashPercentage = 0
	u.ParallelUploads = 0

	u.FailFast = true
	u.Progress = newLoggingUploadProgress(log(ctx))

	return u
}

func onCtrlC(f func()) {
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)

	go func() {
		// invoke the function when Ctrl-C signal is delivered
		<-s
		f()
	}()
}

func dumpSingleFile(ctx context.Context, rep repo.Repository, snapshotID string, fileName string, offset, length int64) (io.ReadCloser, error) {
	rootEntry, err := snapshotfs.FilesystemEntryFromIDWithPath(ctx, rep, snapshotID+"/"+fileName, true)
	if err != nil {
		return nil, err
	}
	file, ok := rootEntry.(fs.File)
	if !ok {
		return nil, fmt.Errorf("unexpected type %T, expected fs.File", rootEntry)
	}
	reader, err := file.Open(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to open fs.File, %w", err)
	}
	if offset > 0 {
		_, err := reader.Seek(offset, io.SeekStart)
		if err != nil {
			return nil, fmt.Errorf("unable to seek, offset: %d, err: %w", offset, err)
		}
	}
	if length >= 0 {
		return &struct {
			io.Reader
			io.Closer
		}{
			Reader: io.LimitReader(reader, length),
			Closer: reader,
		}, nil
	}
	return reader, nil
}
