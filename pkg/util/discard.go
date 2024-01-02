package util

import (
	"io"
	"sync"
)

type discardN struct {
	mu      sync.Mutex
	out     io.Writer
	n       int
	skipped int
}

func (d *discardN) Write(p []byte) (int, error) {
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

func DiscardN(out io.Writer, n int) io.Writer {
	return &discardN{
		out: out,
		n:   n,
	}
}
