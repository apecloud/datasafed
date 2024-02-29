package encryption

import (
	"io"
)

// StreamEncryptor represents an interface for encrypting and decrypting data streams.
type StreamEncryptor interface {
	// EncryptStream appends the encrypted bytes corresponding to the given plaintext to a given writer.
	EncryptStream(plainText io.Reader, output io.Writer) error

	// DecryptStream appends the unencrypted bytes corresponding to the given ciphertext to a given writer.
	DecryptStream(cipherText io.Reader, output io.Writer) error

	// Overhead is the number of bytes of overhead added by EncryptStream()
	Overhead() int
}
