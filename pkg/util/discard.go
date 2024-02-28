package util

import (
	"io"
	"sync"
)

type discardNWriter struct {
	mu      sync.Mutex
	out     io.Writer
	n       int
	skipped int
}

func (d *discardNWriter) Write(p []byte) (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	rest := d.n - d.skipped
	if rest > 0 {
		if len(p) > rest {
			d.skipped += rest
			n, err := d.out.Write(p[rest:])
			return n + rest, err
		} else {
			d.skipped += len(p)
			return len(p), nil
		}
	}
	return d.out.Write(p)
}

func DiscardNWriter(out io.Writer, n int) io.Writer {
	return &discardNWriter{
		out: out,
		n:   n,
	}
}

type discardNReader struct {
	mu      sync.Mutex
	in      io.Reader
	n       int
	skipped int
}

func (d *discardNReader) discardLocked() error {
	buf := make([]byte, 32*1024)
	for {
		rest := d.n - d.skipped
		if rest <= 0 {
			break
		}
		rn := len(buf)
		if rn > rest {
			rn = rest
		}
		n, err := d.in.Read(buf[:rn])
		if err != nil {
			return err
		}
		if n > 0 {
			d.skipped += n
		}
	}
	return nil
}

func (d *discardNReader) Read(p []byte) (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.n-d.skipped > 0 {
		err := d.discardLocked()
		if err != nil {
			return 0, err
		}
	}
	return d.in.Read(p)
}

func DiscardNReader(in io.Reader, n int) io.Reader {
	return &discardNReader{
		in: in,
		n:  n,
	}
}
