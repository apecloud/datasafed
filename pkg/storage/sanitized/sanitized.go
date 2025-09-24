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

var errInvalidPath = errors.New("invalid path")

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

func pathError(format string, args ...any) error {
	return fmt.Errorf("%w: %s", errInvalidPath, fmt.Sprintf(format, args...))
}

func verifiedBasePath(basePath string) (string, error) {
	if basePath == "" {
		return "", nil
	}
	basePath = filepath.Clean(basePath)
	if strings.HasPrefix(basePath, "..") {
		return "", pathError("base path %q prefixed with '..'", basePath)
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
	if rpath == ".." || strings.HasPrefix(rpath, "../") {
		return "", fmt.Errorf("prefixed with '..'")
	}
	if s.basePath == "" {
		return rpath, nil
	}
	return filepath.Join(s.basePath, rpath), nil
}

func (s *sanitizedStorage) Push(ctx context.Context, r io.Reader, rpath string) error {
	if strings.HasSuffix(rpath, "/") {
		return pathError("rpath %q ends with '/'", rpath)
	}
	relocatedPath, err := s.relocate(rpath)
	if err != nil {
		return pathError("invalid rpath %q: %s", rpath, err)
	}
	return s.underlying.Push(ctx, r, relocatedPath)
}

func (s *sanitizedStorage) Pull(ctx context.Context, rpath string, w io.Writer) error {
	if strings.HasSuffix(rpath, "/") {
		return pathError("rpath %q ends with '/'", rpath)
	}
	relocatedPath, err := s.relocate(rpath)
	if err != nil {
		return pathError("invalid rpath %q: %s", rpath, err)
	}
	return s.underlying.Pull(ctx, relocatedPath, w)
}

func (s *sanitizedStorage) Remove(ctx context.Context, rpath string, recursive bool) error {
	relocatedPath, err := s.relocate(rpath)
	if err != nil {
		return pathError("invalid rpath %q: %s", rpath, err)
	}
	return s.underlying.Remove(ctx, relocatedPath, recursive)
}

func (s *sanitizedStorage) Rmdir(ctx context.Context, rpath string) error {
	relocatedPath, err := s.relocate(rpath)
	if err != nil {
		return pathError("invalid rpath %q: %s", rpath, err)
	}
	return s.underlying.Rmdir(ctx, relocatedPath)
}

func (s *sanitizedStorage) Mkdir(ctx context.Context, rpath string) error {
	relocatedPath, err := s.relocate(rpath)
	if err != nil {
		return pathError("invalid rpath %q: %s", rpath, err)
	}
	return s.underlying.Mkdir(ctx, relocatedPath)
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
		return pathError("rpath %q ends with '/', but PathIsFile is true", rpath)
	}
	originalRpath := rpath
	isDir := strings.HasSuffix(rpath, "/")
	relocatedPath, err := s.relocate(rpath)
	if err != nil {
		return pathError("invalid rpath %q: %s", rpath, err)
	}
	if isDir {
		relocatedPath += "/"
	}
	myCb := func(e storage.DirEntry) error {
		return cb(s.adjustPath(ctx, e))
	}
	err = s.underlying.List(ctx, relocatedPath, opt, myCb)
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
	relocatedPath, err := s.relocate(rpath)
	if err != nil {
		return storage.StatResult{}, pathError("invalid rpath %q: %s", rpath, err)
	}
	if isDir {
		relocatedPath += "/"
	}
	return s.underlying.Stat(ctx, relocatedPath)
}

func (s *sanitizedStorage) OpenFile(ctx context.Context, rpath string, offset int64, length int64) (io.ReadCloser, error) {
	if strings.HasSuffix(rpath, "/") {
		return nil, pathError("rpath %q ends with '/'", rpath)
	}
	relocatedPath, err := s.relocate(rpath)
	if err != nil {
		return nil, pathError("invalid rpath %q: %s", rpath, err)
	}
	return s.underlying.OpenFile(ctx, relocatedPath, offset, length)
}

func (s *sanitizedStorage) Unwrap() storage.Storage {
	return s.underlying
}
