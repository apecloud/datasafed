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

type Storage interface {
	Push(ctx context.Context, r io.Reader, rpath string) error
	Pull(ctx context.Context, rpath string, w io.Writer) error
	Remove(ctx context.Context, rpath string, recursive bool) error
	Rmdir(ctx context.Context, rpath string) error
	List(ctx context.Context, rpath string, opt *ListOptions) ([]DirEntry, error)
	Stat(ctx context.Context, rpath string) (StatResult, error)
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
