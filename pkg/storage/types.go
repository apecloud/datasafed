package storage

import (
	"context"
	"io"
	"time"
)

type DirEntry interface {
	IsDir() bool
	Name() string
	Path() string
	Size() int64
	MTime() time.Time
}

type ListOptions struct {
	DirsOnly  bool
	FilesOnly bool
	MaxDepth  int
	Recursive bool
}

type StatResult struct {
	TotalSize int64
	Entries   int64
	Dirs      int64
	Files     int64
}

// Storage is the interface that wraps the basic operations of a storage.
// The Storage interface can also be used to implement common features
// such as data encryption and compression. Multiple storage implementations
// can be linked together as a stack to form a composite storage.
type Storage interface {
	// Push pushes the contents of the reader to the given path.
	// If the path exists, it will be overwritten.
	Push(ctx context.Context, r io.Reader, rpath string) error

	// Pull pulls the contents of the given path to the writer.
	Pull(ctx context.Context, rpath string, w io.Writer) error

	// Remove removes the file in the given path if recursive is false.
	// Otherwise the `rpath` parameter can be a directory and all the
	// contents inside will be removed.
	Remove(ctx context.Context, rpath string, recursive bool) error

	// Rmdir removes the directory in the given path only if the directory is empty.
	Rmdir(ctx context.Context, rpath string) error

	// List lists the contents of the given path.
	// The `rpath` parameter can also be a file, in this case the function
	// will return a list with a single entry.
	List(ctx context.Context, rpath string, opt *ListOptions) ([]DirEntry, error)

	// Stat returns the information about the given path.
	// The `rpath` parameter can be a file.
	Stat(ctx context.Context, rpath string) (StatResult, error)

	// ReadObject reads a file.
	ReadObject(ctx context.Context, rpath string) (io.ReadCloser, error)
}

type staticDirEntry struct {
	isDir bool
	name  string
	path  string
	size  int64
	mtime time.Time
}

func (e *staticDirEntry) IsDir() bool      { return e.isDir }
func (e *staticDirEntry) Name() string     { return e.name }
func (e *staticDirEntry) Path() string     { return e.path }
func (e *staticDirEntry) Size() int64      { return e.size }
func (e *staticDirEntry) MTime() time.Time { return e.mtime }

func NewStaticDirEntry(isDir bool, name, path string, size int64, mtime time.Time) DirEntry {
	return &staticDirEntry{
		isDir: isDir,
		name:  name,
		path:  path,
		size:  size,
		mtime: mtime,
	}
}
