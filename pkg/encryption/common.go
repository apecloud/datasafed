package encryption

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

const minDerivedKeyLength = 32

func pipeStream(in io.Reader, inName string,
	out io.Writer, outName string,
	buf []byte, manipulateFunc func([]byte)) error {
	for {
		n, err := in.Read(buf)
		if n > 0 {
			data := buf[:n]
			manipulateFunc(data)
			wn, err := out.Write(data)
			if err != nil {
				return fmt.Errorf("write to %s error: %w", outName, err)
			}
			if wn != n {
				return fmt.Errorf("partially write to %s, length: %d, actual write: %d",
					outName, n, wn)
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			} else {
				return fmt.Errorf("read from %s error: %w", inName, err)
			}
		}
	}
}

// deriveKey uses HKDF to derive a key of a given length and a given purpose from parameters.
func deriveKey(passPhrase []byte, purpose []byte, length int) ([]byte, error) {
	if length < minDerivedKeyLength {
		return nil, fmt.Errorf("derived key must be at least 32 bytes, was %v", length)
	}

	key := make([]byte, length)
	k := hkdf.New(sha256.New, passPhrase, purpose, nil)
	_, err := io.ReadFull(k, key)
	if err != nil {
		return nil, err
	}

	return key, nil
}
