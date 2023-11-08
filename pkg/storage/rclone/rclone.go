package rclone

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	// TODO: exclude net-disk products to reduce the binary size
	_ "github.com/rclone/rclone/backend/all"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config"
	"github.com/rclone/rclone/fs/operations"

	"github.com/apecloud/datasafed/pkg/storage"
)

const (
	remoteName = "backend"
	rootKey    = "root"
)

type rcloneStorage struct {
	f fs.Fs
}

var _ storage.Storage = (*rcloneStorage)(nil)

func New(cfg map[string]string) (storage.Storage, error) {
	rcloneCfg := config.Data()
	for k, v := range cfg {
		rcloneCfg.SetValue(remoteName, k, v)
	}
	root := cfg[rootKey]
	f, err := fs.NewFs(context.Background(), remoteName+":"+root)
	if err != nil {
		return nil, err
	}
	return &rcloneStorage{
		f: f,
	}, nil
}

func (s *rcloneStorage) Push(ctx context.Context, r io.Reader, rpath string) error {
	rpath = normalizeRemotePath(rpath)
	// check if rpath is a directory
	_, err := s.f.NewObject(ctx, rpath)
	if errors.Is(err, fs.ErrorIsDir) {
		return err
	}
	// use RcatSize() to upload if it's a regular file
	if f, ok := r.(*os.File); ok {
		fi, err := f.Stat()
		if err == nil && fi.Mode().IsRegular() {
			_, err = operations.RcatSize(ctx, s.f, rpath, io.NopCloser(r), fi.Size(), time.Now(), nil)
			return err
		}
	}
	// streaming upload with Rcat()
	if s.f.Features().PutStream == nil {
		fmt.Fprintln(os.Stderr, "Warning: target remote doesn't support streaming uploads,"+
			" may save the content to a temporary file before uploading")
	}
	_, err = operations.Rcat(ctx, s.f, rpath, io.NopCloser(r), time.Now(), nil)
	return err
}

func (s *rcloneStorage) Pull(ctx context.Context, rpath string, w io.Writer) error {
	rpath = normalizeRemotePath(rpath)
	obj, err := s.f.NewObject(ctx, rpath)
	if err != nil {
		return err
	}
	rc, err := obj.Open(ctx)
	if err != nil {
		return err
	}
	defer rc.Close()
	_, err = io.Copy(w, rc)
	return err
}

func (s *rcloneStorage) ReadObject(ctx context.Context, rpath string) (io.ReadCloser, error) {
	rpath = normalizeRemotePath(rpath)
	obj, err := s.f.NewObject(ctx, rpath)
	if err != nil {
		return nil, err
	}
	return obj.Open(ctx)
}

func (s *rcloneStorage) Remove(ctx context.Context, rpath string, recursive bool) error {
	rpath = normalizeRemotePath(rpath)
	if !recursive {
		obj, err := s.f.NewObject(ctx, rpath)
		if err != nil {
			if errors.Is(err, fs.ErrorObjectNotFound) {
				return nil
			}
			return err
		}
		return obj.Remove(ctx)
	} else {
		return operations.Purge(ctx, s.f, rpath)
	}
}

func (s *rcloneStorage) Rmdir(ctx context.Context, rpath string) error {
	rpath = normalizeRemotePath(rpath)
	return s.f.Rmdir(ctx, rpath)
}

func (s *rcloneStorage) Mkdir(ctx context.Context, rpath string) error {
	rpath = normalizeRemotePath(rpath)
	return s.f.Mkdir(ctx, rpath)
}

func (s *rcloneStorage) list(ctx context.Context, rpath string, opt *storage.ListOptions, callback func(item *operations.ListJSONItem) error) error {
	ljOpt := &operations.ListJSONOpt{}
	if opt != nil {
		ljOpt.Recurse = opt.Recursive
		ljOpt.DirsOnly = opt.DirsOnly
		ljOpt.FilesOnly = opt.FilesOnly
		if opt.MaxDepth > 0 {
			var ci *fs.ConfigInfo
			ctx, ci = fs.AddConfig(ctx)
			ci.MaxDepth = opt.MaxDepth
		}
	}
	return operations.ListJSON(ctx, s.f, rpath, ljOpt, callback)
}

func (s *rcloneStorage) List(ctx context.Context, rpath string, opt *storage.ListOptions) ([]storage.DirEntry, error) {
	rpath = normalizeRemotePath(rpath)
	obj, err := s.f.NewObject(ctx, rpath)
	if err == nil {
		entry := storage.NewStaticDirEntry(false, filepath.Base(obj.Remote()),
			obj.Remote(), obj.Size(), obj.ModTime(ctx))
		return []storage.DirEntry{entry}, nil
	}
	var result []storage.DirEntry
	err = s.list(ctx, rpath, opt, func(item *operations.ListJSONItem) error {
		en := storage.NewStaticDirEntry(item.IsDir, item.Name, item.Path, item.Size, item.ModTime.When)
		result = append(result, en)
		return nil
	})
	return result, err
}

func (s *rcloneStorage) Stat(ctx context.Context, rpath string) (storage.StatResult, error) {
	rpath = normalizeRemotePath(rpath)
	obj, err := s.f.NewObject(ctx, rpath)
	if err == nil {
		return storage.StatResult{
			TotalSize: obj.Size(),
			Entries:   1,
			Files:     1,
		}, nil
	}

	opt := &storage.ListOptions{
		Recursive: true,
	}
	var result storage.StatResult
	err = s.list(ctx, rpath, opt, func(item *operations.ListJSONItem) error {
		if item.IsBucket {
			// ignore buckets
			return nil
		}
		if item.IsDir {
			result.Dirs++
		} else {
			result.Files++
			result.TotalSize += item.Size
		}
		return nil
	})
	result.Entries = result.Dirs + result.Files
	return result, err
}

func normalizeRemotePath(rpath string) string {
	rpath = filepath.Clean(rpath)
	rpath = strings.TrimPrefix(rpath, "./")
	rpath = strings.TrimPrefix(rpath, "/")
	if rpath == "." {
		// rclone doesn't accept "." as a remote path
		return ""
	}
	return rpath
}
