package kopia

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/apecloud/datasafed/pkg/storage"
	"github.com/kopia/kopia/repo/blob"
	"github.com/kopia/kopia/repo/blob/sharded"
)

const (
	storageType = "datasafed"
)

func init() {
	blob.AddSupportedStorage(storageType, blobOptions{}, newBlobStorage)
}

type blobOptions struct {
	RootPath string `json:"root_path"`
	Caching  bool   `json:"caching"`
}

type blobStorageImpl struct {
	sharded.Storage
	blob.DefaultProviderImplementation

	s storage.Storage
}

var _ blob.Storage = (*blobStorageImpl)(nil)
var _ sharded.Impl = (*blobStorageImpl)(nil)

func newBlobStorage(ctx context.Context, options *blobOptions, isCreate bool) (blob.Storage, error) {
	underlying := GetUnderlyingStorage()
	if underlying == nil {
		return nil, fmt.Errorf("SetUnderlyingStorage() should be called first")
	}
	impl := &blobStorageImpl{s: underlying}
	impl.Storage = sharded.New(impl, options.RootPath, sharded.Options{ListParallelism: 8}, isCreate)
	return impl, nil
}

func (b *blobStorageImpl) GetBlobFromPath(ctx context.Context, dirPath, filePath string, offset, length int64, output blob.OutputBuffer) error {
	log(ctx).Debugf("GetBlobFromPath dir %s file %s, offset %d length %d", dirPath, filePath, offset, length)
	rd, err := b.s.OpenFile(ctx, filePath, offset, length)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotFound) {
			return blob.ErrBlobNotFound
		}
		return fmt.Errorf("fail to OpenFile, err: %w", err)
	}
	defer rd.Close()
	_, err = io.Copy(output, rd)
	if err != nil {
		return fmt.Errorf("fail to copy, err: %w", err)
	}
	return blob.EnsureLengthExactly(output.Length(), length)
}

func (b *blobStorageImpl) GetMetadataFromPath(ctx context.Context, dirPath, filePath string) (blob.Metadata, error) {
	log(ctx).Debugf("GetMetadata dir %s file %s", dirPath, filePath)
	entries, err := b.s.List(ctx, filePath, &storage.ListOptions{PathIsFile: true})
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotFound) {
			return blob.Metadata{}, blob.ErrBlobNotFound
		}
		return blob.Metadata{}, fmt.Errorf("fail to OpenFile, err: %w", err)
	}
	if len(entries) != 1 {
		return blob.Metadata{}, fmt.Errorf("expect List() to return single entry for path %s, got %d", filePath, len(entries))
	}
	entry := entries[0]
	return blob.Metadata{
		Length:    entry.Size(),
		Timestamp: entry.MTime(),
	}, nil
}

func (b *blobStorageImpl) PutBlobInPath(ctx context.Context, dirPath, filePath string, dataSlices blob.Bytes, opts blob.PutOptions) error {
	log(ctx).Debugf("PutBlobInPath dir %s file %s", dirPath, filePath)
	switch {
	case opts.HasRetentionOptions():
		return fmt.Errorf("%w: blob-retention", blob.ErrUnsupportedPutBlobOption)
	case opts.DoNotRecreate:
		return fmt.Errorf("%w: do-not-recreate", blob.ErrUnsupportedPutBlobOption)
	case !opts.SetModTime.IsZero():
		return blob.ErrSetTimeUnsupported
	}

	// according to this post https://forum.rclone.org/t/are-transfers-atomic/21704/6 ,
	// the transfer is atomic for most of the backends supported by rclone.
	err := b.s.Push(ctx, dataSlices.Reader(), filePath)

	if opts.GetModTime != nil {
		bm, err2 := b.GetMetadataFromPath(ctx, dirPath, filePath)
		if err2 != nil {
			return err2
		}

		*opts.GetModTime = bm.Timestamp
	}
	return err
}

func (b *blobStorageImpl) DeleteBlobInPath(ctx context.Context, dirPath, filePath string) error {
	log(ctx).Debugf("DeleteBlobInPath dir %s file %s", dirPath, filePath)
	return b.s.Remove(ctx, filePath, false)
}

func (b *blobStorageImpl) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	log(ctx).Debugf("ReadDir path %s", path)
	var fileInfos []os.FileInfo
	_, err := b.s.List(ctx, path, &storage.ListOptions{
		Recursive: false,
		Callback: func(en storage.DirEntry) error {
			info := fileInfo{
				name:    en.Name(),
				size:    en.Size(),
				modTime: en.MTime(),
				isDir:   en.IsDir(),
			}
			fileInfos = append(fileInfos, &info)
			return nil
		},
	})
	return fileInfos, err
}

func (b *blobStorageImpl) ConnectionInfo() blob.ConnectionInfo {
	return blob.ConnectionInfo{
		Type: storageType,
		Config: blobOptions{
			RootPath: b.RootPath,
		},
	}
}

func (b *blobStorageImpl) DisplayName() string {
	return storageType
}

type fileInfo struct {
	name    string
	size    int64
	modTime time.Time
	isDir   bool
}

func (f *fileInfo) Name() string       { return f.name }
func (f *fileInfo) Size() int64        { return f.size }
func (f *fileInfo) ModTime() time.Time { return f.modTime }
func (f *fileInfo) IsDir() bool        { return f.isDir }
func (f *fileInfo) Sys() any           { return nil }

func (f *fileInfo) Mode() os.FileMode {
	if f.isDir {
		return 0755 | os.ModeDir
	} else {
		return 0644
	}
}
