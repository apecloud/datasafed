package encrypted

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/apecloud/datasafed/pkg/encryption"
	"github.com/apecloud/datasafed/pkg/logging"
	"github.com/apecloud/datasafed/pkg/storage"
	"github.com/apecloud/datasafed/pkg/storage/sanitized"
	"github.com/apecloud/datasafed/pkg/util"
)

const encryptedFileSuffix = ".enc"

var log = logging.Module("storage/encrypted")

type encryptedStorage struct {
	encryptor  encryption.StreamEncryptor
	underlying storage.Storage
}

var _ storage.Storage = (*encryptedStorage)(nil)

func New(ctx context.Context,
	encryptor encryption.StreamEncryptor,
	underlying storage.Storage) (storage.Storage, error) {
	es := &encryptedStorage{
		encryptor:  encryptor,
		underlying: underlying,
	}
	return sanitized.New(ctx, "", es)
}

func (s *encryptedStorage) Push(ctx context.Context, r io.Reader, rpath string) error {
	pr, pw := io.Pipe()
	go func() {
		err := s.encryptor.EncryptStream(r, pw)
		if err != nil {
			pw.CloseWithError(err)
		} else {
			pw.Close() // EOF
		}
	}()
	return s.underlying.Push(ctx, pr, rpath+encryptedFileSuffix)
}

func (s *encryptedStorage) Pull(ctx context.Context, rpath string, w io.Writer) error {
	errCh := make(chan error, 1)
	pr, pw := io.Pipe()
	go func() {
		err := s.encryptor.DecryptStream(pr, w)
		if err != nil {
			// interrupt underlying.Pull()
			pr.CloseWithError(err)
		}
		errCh <- err
	}()
	err := s.underlying.Pull(ctx, rpath+encryptedFileSuffix, pw)
	if err != nil {
		pw.CloseWithError(err)
	} else {
		pw.Close()
	}
	decErr := <-errCh // wait until all data are decrypted
	if err != nil {
		return err
	}
	return decErr
}

func (s *encryptedStorage) OpenFile(ctx context.Context, rpath string, offset int64, length int64) (io.ReadCloser, error) {
	rc, err := s.underlying.OpenFile(ctx, rpath+encryptedFileSuffix, 0, 0)
	if err != nil {
		return nil, err
	}
	pr, pw := io.Pipe()
	go func() {
		err := s.encryptor.DecryptStream(rc, pw)
		if err != nil {
			pw.CloseWithError(err)
		} else {
			pw.Close()
		}
	}()
	var rd io.Reader = pr
	if offset > 0 {
		log(ctx).Warnf("[ENCRYPTED] OpenFile(): it's not efficient to handle non-zero offset(%d)", offset)
		rd = util.DiscardNReader(rd, int(offset))
	}
	if length > 0 {
		rd = io.LimitReader(rd, length)
	}
	return &struct {
		io.Reader
		io.Closer
	}{
		Reader: rd,
		Closer: pr,
	}, nil
}

func (s *encryptedStorage) Remove(ctx context.Context, rpath string, recursive bool) error {
	if !recursive {
		return s.underlying.Remove(ctx, rpath+encryptedFileSuffix, false)
	} else {
		return s.underlying.Remove(ctx, rpath, true)
	}
}

func (s *encryptedStorage) Rmdir(ctx context.Context, rpath string) error {
	return s.underlying.Rmdir(ctx, rpath)
}

func (s *encryptedStorage) Mkdir(ctx context.Context, rpath string) error {
	return s.underlying.Mkdir(ctx, rpath)
}

func (s *encryptedStorage) List(ctx context.Context, rpath string, opt *storage.ListOptions, cb storage.ListCallback) error {
	myCb := func(de storage.DirEntry) error {
		var err error
		if !de.IsDir() {
			if strings.HasSuffix(de.Name(), encryptedFileSuffix) {
				name := strings.TrimSuffix(de.Name(), encryptedFileSuffix)
				path := strings.TrimSuffix(de.Path(), encryptedFileSuffix)
				size := de.Size() - int64(s.encryptor.Overhead())
				newEntry := storage.NewStaticDirEntry(de.IsDir(), name, path, size, de.MTime())
				err = cb(newEntry)
			}
			// ignore files that doesn't end with encryptedFileSuffix
		} else {
			err = cb(de)
		}
		return err
	}
	if opt.PathIsFile {
		// list a file
		return s.underlying.List(ctx, rpath+encryptedFileSuffix, opt, myCb)
	} else if strings.HasSuffix(rpath, "/") || rpath == "." {
		// rpath is a folder
		return s.underlying.List(ctx, rpath, opt, myCb)
	} else {
		// try list single file first
		cloneOpt := *opt
		cloneOpt.PathIsFile = true
		err := s.underlying.List(ctx, rpath+encryptedFileSuffix, &cloneOpt, myCb)
		if err != nil {
			// ignore ErrObjectNotFound
			if !errors.Is(err, storage.ErrObjectNotFound) {
				return err
			}
		} else {
			return nil
		}

		// try list a folder
		return s.underlying.List(ctx, rpath, opt, myCb)
	}
}

func (s *encryptedStorage) Stat(ctx context.Context, rpath string) (storage.StatResult, error) {
	result := storage.StatResult{}
	statFunc := func(de storage.DirEntry) error {
		if de.IsDir() {
			result.Dirs++
		} else {
			result.Files++
			result.TotalSize += de.Size()
		}
		return nil
	}
	err := s.List(ctx, rpath, &storage.ListOptions{Recursive: true}, statFunc)
	result.Entries = result.Dirs + result.Files
	return result, err
}
