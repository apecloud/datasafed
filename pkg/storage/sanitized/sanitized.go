package sanitized

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/apecloud/datasafed/pkg/logging"
	"github.com/apecloud/datasafed/pkg/storage"
)

var log = logging.Module("storage/sanitized")

type sanitizedStorage struct {
	basePath   string
	underlying storage.Storage
}

func New(ctx context.Context, basePath string, underlying storage.Storage) (storage.Storage, error) {
	basePath, err := verifiedBasePath(basePath)
	if err != nil {
		return nil, err
	}
	return &sanitizedStorage{
		basePath:   basePath,
		underlying: underlying,
	}, nil
}

func verifiedBasePath(basePath string) (string, error) {
	if basePath == "" {
		return "", nil
	}
	basePath = filepath.Clean(basePath)
	if strings.HasPrefix(basePath, "..") {
		return "", fmt.Errorf("base path %q prefixed with '..'", basePath)
	}
	if basePath == "." {
		basePath = ""
	} else {
		basePath = strings.TrimPrefix(basePath, "/")
		basePath = strings.TrimPrefix(basePath, "./")
	}
	return basePath, nil
}

func (s *sanitizedStorage) relocate(rpath string) (string, error) {
	rpath = filepath.Clean(rpath)
	rpath = strings.TrimPrefix(rpath, "./")
	rpath = strings.TrimPrefix(rpath, "/")
	if strings.HasPrefix(rpath, "..") {
		return "", fmt.Errorf("prefixed with '..'")
	}
	if s.basePath == "" {
		return rpath, nil
	}
	return filepath.Join(s.basePath, rpath), nil
}

func (s *sanitizedStorage) Push(ctx context.Context, r io.Reader, rpath string) error {
	if strings.HasSuffix(rpath, "/") {
		return fmt.Errorf("rpath %q ends with '/'", rpath)
	}
	rpath, err := s.relocate(rpath)
	if err != nil {
		return fmt.Errorf("invalid rpath %q: %w", rpath, err)
	}
	return s.underlying.Push(ctx, r, rpath)
}

func (s *sanitizedStorage) Pull(ctx context.Context, rpath string, w io.Writer) error {
	if strings.HasSuffix(rpath, "/") {
		return fmt.Errorf("rpath %q ends with '/'", rpath)
	}
	rpath, err := s.relocate(rpath)
	if err != nil {
		return fmt.Errorf("invalid rpath %q: %w", rpath, err)
	}
	return s.underlying.Pull(ctx, rpath, w)
}

func (s *sanitizedStorage) Remove(ctx context.Context, rpath string, recursive bool) error {
	rpath, err := s.relocate(rpath)
	if err != nil {
		return fmt.Errorf("invalid rpath %q: %w", rpath, err)
	}
	return s.underlying.Remove(ctx, rpath, recursive)
}

func (s *sanitizedStorage) Rmdir(ctx context.Context, rpath string) error {
	rpath, err := s.relocate(rpath)
	if err != nil {
		return fmt.Errorf("invalid rpath %q: %w", rpath, err)
	}
	return s.underlying.Rmdir(ctx, rpath)
}

func (s *sanitizedStorage) Mkdir(ctx context.Context, rpath string) error {
	rpath, err := s.relocate(rpath)
	if err != nil {
		return fmt.Errorf("invalid rpath %q: %w", rpath, err)
	}
	return s.underlying.Mkdir(ctx, rpath)
}

func (s *sanitizedStorage) adjustPath(ctx context.Context, e storage.DirEntry) storage.DirEntry {
	if s.basePath == "" {
		return e
	}
	final, err := filepath.Rel(s.basePath, e.Path())
	if err != nil {
		log(ctx).Warnf("[SANITIZED] failed to get relative path %q to %q: %v", e.Path(), s.basePath, err)
		return e
	}
	return storage.NewStaticDirEntry(e.IsDir(), e.Name(), final, e.Size(), e.MTime())
}

func (s *sanitizedStorage) List(ctx context.Context, rpath string, opt *storage.ListOptions, cb storage.ListCallback) error {
	if opt.PathIsFile && strings.HasSuffix(rpath, "/") {
		return fmt.Errorf("rpath %q ends with '/', but PathIsFile is true", rpath)
	}
	originalRpath := rpath
	isDir := strings.HasSuffix(rpath, "/")
	rpath, err := s.relocate(rpath)
	if err != nil {
		return fmt.Errorf("invalid rpath %q: %w", rpath, err)
	}
	if isDir {
		rpath += "/"
	}
	myCb := func(e storage.DirEntry) error {
		return cb(s.adjustPath(ctx, e))
	}
	err = s.underlying.List(ctx, rpath, opt, myCb)
	if errors.Is(err, storage.ErrDirNotFound) {
		if originalRpath == "/" || originalRpath == "." {
			// ignore directory not found error for root directory
			return nil
		}
	}
	return err
}

func (s *sanitizedStorage) Stat(ctx context.Context, rpath string) (storage.StatResult, error) {
	isDir := strings.HasSuffix(rpath, "/")
	rpath, err := s.relocate(rpath)
	if err != nil {
		return storage.StatResult{}, fmt.Errorf("invalid rpath %q: %w", rpath, err)
	}
	if isDir {
		rpath += "/"
	}
	return s.underlying.Stat(ctx, rpath)
}

func (s *sanitizedStorage) OpenFile(ctx context.Context, rpath string, offset int64, length int64) (io.ReadCloser, error) {
	if strings.HasSuffix(rpath, "/") {
		return nil, fmt.Errorf("rpath %q ends with '/'", rpath)
	}
	rpath, err := s.relocate(rpath)
	if err != nil {
		return nil, fmt.Errorf("invalid rpath %q: %w", rpath, err)
	}
	return s.underlying.OpenFile(ctx, rpath, offset, length)
}

func (s *sanitizedStorage) Unwrap() storage.Storage {
	return s.underlying
}
