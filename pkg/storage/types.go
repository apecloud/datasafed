package storage

import (
	"context"
	"errors"
	"io"
	"time"
)

var (
	ErrObjectNotFound = errors.New("object not found")
	ErrDirNotFound    = errors.New("directory not found")
	ErrIsDir          = errors.New("path is a directory")
)

type DirEntry interface {
	IsDir() bool
	Name() string
	Path() string
	Size() int64
	MTime() time.Time
}

type ListOptions struct {
	DirsOnly   bool
	FilesOnly  bool
	MaxDepth   int
	Recursive  bool
	PathIsFile bool
	Callback   func(DirEntry) error
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

	// Mkdir makes a directory.
	Mkdir(ctx context.Context, rpath string) error

	// List lists the contents of the given path.
	// The `rpath` parameter can also be a file, in this case the function
	// will return a list with a single entry.
	// It's recommended to add '/' to the end of `rpath` if it points to a directory.
	// If a `Callback` is specified in the `ListOptions`, it will be called
	// for each entry. In this case, the method returns an empty list. If the
	// callback returns an error, the function will stop and return the error.
	List(ctx context.Context, rpath string, opt *ListOptions) ([]DirEntry, error)

	// Stat returns the information about the given path.
	// The `rpath` parameter can be a file.
	// It's recommended to add '/' to the end of `rpath` if it points to a directory.
	Stat(ctx context.Context, rpath string) (StatResult, error)

	// OpenFile returns a file reader for the given path.
	// The `offset` and `length` parameters are used to specify the range
	// of the file to read. If `length` is -1, the entire file will be read.
	OpenFile(ctx context.Context, rpath string, offset int64, length int64) (io.ReadCloser, error)
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
