package kopia

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kopia/kopia/repo"
	"github.com/kopia/kopia/repo/manifest"
	"github.com/kopia/kopia/snapshot"

	"github.com/apecloud/datasafed/pkg/logging"
	"github.com/apecloud/datasafed/pkg/storage"
	"github.com/apecloud/datasafed/pkg/storage/sanitized"
	"github.com/apecloud/datasafed/pkg/util"
)

const (
	RepoRootKey     = "kopia.repo_root"
	PasswordKey     = "kopia.password"
	DisableCacheKey = "kopia.disable_cache"

	metaSuffix = ".meta"
	tmpSuffix  = ".tmp"
)

var log = logging.Module("storage/kopia")

type meta struct {
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	ModTime    time.Time `json:"mod_time"`
	SnapshotID string    `json:"snapshot_id"`
}

type kopiaStorage struct {
	rep        repo.Repository
	underlying storage.Storage
}

var _ storage.Storage = (*kopiaStorage)(nil)

func New(ctx context.Context, cfg map[string]string, basePath string) (storage.Storage, error) {
	underlying := GetUnderlyingStorage()
	if underlying == nil {
		return nil, fmt.Errorf("SetUnderlyingStorage() should be called first")
	}

	repoRootPath := cfg[RepoRootKey]
	repoRootPath = filepath.Clean(repoRootPath)
	if repoRootPath == "." || repoRootPath == "/" || strings.HasPrefix(repoRootPath, "..") {
		return nil, fmt.Errorf("kopia repo root path should not be '.', '/' or started with '..', path: %q", repoRootPath)
	}

	// after filepath.Clean(), repoRootPath should not contain '/' at the end
	repoMetaPath := repoRootPath + metaSuffix
	underlying, err := sanitized.New(ctx, repoMetaPath, underlying)
	if err != nil {
		return nil, fmt.Errorf("sanitized.New error: %w", err)
	}

	cacheID := ""
	if disabled, _ := strconv.ParseBool(cfg[DisableCacheKey]); !disabled {
		cacheID = generateUniqueID(cfg, basePath)
	}

	password := cfg[PasswordKey]
	rep, err := getInitedRepository(ctx, repoRootPath, password, cacheID)
	if err != nil {
		return nil, fmt.Errorf("getInitedRepository error: %w", err)
	}
	s := &kopiaStorage{
		rep:        rep,
		underlying: underlying,
	}
	return sanitized.New(ctx, basePath, s)
}

func generateUniqueID(cfg map[string]string, basePath string) string {
	data, _ := json.Marshal(cfg)
	hash := md5.New()
	hash.Write(data)
	hash.Write([]byte(basePath))
	return hex.EncodeToString(hash.Sum(nil))
}

func (s *kopiaStorage) Push(ctx context.Context, r io.Reader, rpath string) error {
	log(ctx).Infof("[KOPIA] Push %s", rpath)

	fileName := filepath.Base(rpath)
	if fileName == "" || fileName == "." {
		return fmt.Errorf("invalid file name: %q", fileName)
	}

	// save the old meta file if exists
	var oldMetaFile string
	content, err := s.underlying.OpenFile(ctx, rpath+metaSuffix, 0, -1)
	if err == nil {
		defer content.Close()
		oldMetaFile = rpath + tmpSuffix
		err := s.underlying.Push(ctx, content, oldMetaFile+metaSuffix)
		if err != nil {
			return fmt.Errorf("unable to save old meta file: %w", err)
		}
	} else {
		if !errors.Is(err, storage.ErrObjectNotFound) {
			return fmt.Errorf("unable to open meta file: %w", err)
		}
	}

	// add the file to the kopia repo
	var manifest *snapshot.Manifest
	err = repo.WriteSession(ctx, s.rep, repo.WriteSessionOptions{
		Purpose: "datasafed:push",
	}, func(ctx context.Context, w repo.RepositoryWriter) error {
		var err error
		manifest, err = snapshotSingleFile(ctx, fileName, r, w, nil)
		return err
	})
	if err != nil {
		return err
	}

	// write meta file
	meta := &meta{
		Name:       fileName,
		Size:       manifest.Stats.TotalFileSize,
		ModTime:    manifest.EndTime.ToTime(),
		SnapshotID: string(manifest.ID),
	}
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	if err = enc.Encode(meta); err != nil {
		return fmt.Errorf("marshal meta json failed, meta: %+v, err: %w", meta, err)
	}
	err = s.underlying.Push(ctx, buf, rpath+metaSuffix)
	if err != nil {
		return err
	}

	// remove the shadowed file from the kopia repo
	if oldMetaFile != "" {
		err := s.Remove(ctx, oldMetaFile, false)
		if err != nil {
			log(ctx).Warnf("unable to remove the shadowed file: %v", err)
			return nil
		}
	}

	return nil
}

func (s *kopiaStorage) Pull(ctx context.Context, rpath string, w io.Writer) error {
	log(ctx).Infof("[KOPIA] Pull %s", rpath)

	rc, err := s.OpenFile(ctx, rpath, 0, -1)
	if err != nil {
		return fmt.Errorf("fail to OpenFile, %w", err)
	}
	defer rc.Close()
	_, err = io.Copy(w, rc)
	return err
}

func (s *kopiaStorage) OpenFile(ctx context.Context, rpath string, offset, length int64) (io.ReadCloser, error) {
	log(ctx).Infof("[KOPIA] OpenFile %s", rpath)
	meta, err := s.loadMeta(ctx, rpath)
	if err != nil {
		return nil, err
	}
	rc, err := dumpSingleFile(ctx, s.rep, meta.SnapshotID, meta.Name, offset, length)
	if err != nil {
		return nil, fmt.Errorf("dumpSingleFile error: %w", err)
	}
	return rc, err
}

func (s *kopiaStorage) loadMeta(ctx context.Context, rpath string) (*meta, error) {
	buf := bytes.NewBuffer(nil)
	err := s.underlying.Pull(ctx, rpath+metaSuffix, buf)
	if err != nil {
		return nil, fmt.Errorf("fail to pull underlying %q, %w", rpath+metaSuffix, err)
	}
	meta := &meta{}
	err = json.Unmarshal(buf.Bytes(), meta)
	if err != nil {
		return nil, fmt.Errorf("unmarshal meta json failed, err: %w", err)
	}
	return meta, nil
}

func (s *kopiaStorage) Remove(ctx context.Context, rpath string, recursive bool) error {
	log(ctx).Infof("[KOPIA] Remove %s, recursive: %v", rpath, recursive)

	if !recursive {
		meta, err := s.loadMeta(ctx, rpath)
		if err != nil {
			return err
		}
		err = repo.WriteSession(ctx, s.rep, repo.WriteSessionOptions{
			Purpose: "datasafed:remove",
		}, func(ctx context.Context, w repo.RepositoryWriter) error {
			return w.DeleteManifest(ctx, manifest.ID(meta.SnapshotID))
		})
		if err != nil {
			return fmt.Errorf("fail to remove kopia snapshot %s, error: %w", meta.SnapshotID, err)
		}
		err = s.underlying.Remove(ctx, rpath+metaSuffix, false)
		return util.WrappedErrOrNil(err, "fail to remove underlying %q", rpath+metaSuffix)
	}

	dirStack := []string{rpath}

	// list the dir to remove files and record dirs in order
	if _, err := s.underlying.List(ctx, rpath, &storage.ListOptions{
		Recursive: true,
		Callback: func(en storage.DirEntry) error {
			log(ctx).Debugf("listing %q, is dir: %v", en.Path(), en.IsDir())
			if en.IsDir() {
				dirStack = append(dirStack, en.Path())
				return nil
			}
			// remove the current listed file
			if strings.HasSuffix(en.Name(), metaSuffix) {
				path := strings.TrimSuffix(en.Path(), metaSuffix)
				err := s.Remove(ctx, path, false)
				return util.WrappedErrOrNil(err, "fail to remove %q", path)
			} else {
				log(ctx).Warnf("listing non meta file %q", en.Path())
				return nil
			}
		},
	}); err != nil {
		return err
	}

	// remove dirs in the reversed order
	for i := len(dirStack) - 1; i >= 0; i-- {
		err := s.Rmdir(ctx, dirStack[i])
		if err != nil {
			// ignore the error
			log(ctx).Errorf("fail to rmdir %q, error: %v", dirStack[i], err)
		}
	}
	return nil
}

func (s *kopiaStorage) Rmdir(ctx context.Context, rpath string) error {
	log(ctx).Infof("[KOPIA] Rmdir %s", rpath)
	return s.underlying.Rmdir(ctx, rpath)
}

func (s *kopiaStorage) Mkdir(ctx context.Context, rpath string) error {
	log(ctx).Infof("[KOPIA] Mkdir %s", rpath)
	return s.underlying.Mkdir(ctx, rpath)
}

func (s *kopiaStorage) List(ctx context.Context, rpath string, opt *storage.ListOptions) ([]storage.DirEntry, error) {
	log(ctx).Infof("[KOPIA] List %s, options: %+v", rpath, opt)

	var err error
	if !strings.HasSuffix(rpath, "/") {
		meta, err := s.loadMeta(ctx, rpath)
		if err == nil {
			en := storage.NewStaticDirEntry(false, filepath.Base(rpath), rpath, meta.Size, meta.ModTime)
			if opt.Callback != nil {
				err := opt.Callback(en)
				if err != nil {
					return nil, fmt.Errorf("interrupted by error: %w", err)
				}
				return nil, nil
			}
			return []storage.DirEntry{en}, nil
		}
	}

	if opt.PathIsFile {
		if strings.HasSuffix(rpath, "/") {
			return nil, storage.ErrIsDir
		}
		return nil, err
	}

	callback := opt.Callback
	var result []storage.DirEntry
	cloneOpt := *opt
	cloneOpt.Callback = func(en storage.DirEntry) error {
		if !en.IsDir() {
			if strings.HasSuffix(en.Name(), metaSuffix) {
				filePath := strings.TrimSuffix(en.Path(), metaSuffix)
				meta, err := s.loadMeta(ctx, filePath)
				if err != nil {
					return fmt.Errorf("load meta for file %q failed: %w", filePath, err)
				}
				fileName := strings.TrimSuffix(en.Name(), metaSuffix)
				en = storage.NewStaticDirEntry(false, fileName, filePath, meta.Size, meta.ModTime)
			} else {
				log(ctx).Warnf("listing non meta file %s", en.Path())
				return nil
			}
		}
		if callback != nil {
			return callback(en)
		} else {
			result = append(result, en)
			return nil
		}
	}
	_, err = s.underlying.List(ctx, rpath, &cloneOpt)
	return result, err
}

func (s *kopiaStorage) Stat(ctx context.Context, rpath string) (storage.StatResult, error) {
	log(ctx).Infof("[KOPIA] Stat %s", rpath)

	meta, err := s.loadMeta(ctx, rpath)
	if err == nil {
		return storage.StatResult{
			TotalSize: meta.Size,
			Entries:   1,
			Files:     1,
		}, nil
	}

	var result storage.StatResult
	opt := &storage.ListOptions{
		Recursive: true,
		Callback: func(en storage.DirEntry) error {
			if en.IsDir() {
				result.Dirs++
			} else {
				result.Files++
				result.TotalSize += en.Size()
			}
			return nil
		},
	}

	_, err = s.List(ctx, rpath, opt)
	result.Entries = result.Dirs + result.Files
	return result, err
}
