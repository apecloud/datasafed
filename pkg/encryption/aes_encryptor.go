package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

const bufferSize = ((128 * 1024) / aes.BlockSize) * aes.BlockSize

type aesEncryptor struct {
	key          []byte
	newEncryptor func(block cipher.Block, iv []byte) cipher.Stream
	newDecryptor func(block cipher.Block, iv []byte) cipher.Stream
}

func (e *aesEncryptor) EncryptStream(plainText io.Reader, output io.Writer) error {
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return fmt.Errorf("rand.Read() error: %w", err)
	}
	n, err := output.Write(iv)
	if err != nil {
		return fmt.Errorf("write IV error: %w", err)
	}
	if n != len(iv) {
		return fmt.Errorf("partially write IV, len: %d, written: %d", len(iv), n)
	}
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return fmt.Errorf("aes.NewCipher() error: %w", err)
	}
	enc := e.newEncryptor(block, iv)
	buf := make([]byte, bufferSize)
	return pipeStream(plainText, "plainText", output, "cipherText", buf, func(b []byte) {
		enc.XORKeyStream(b, b)
	})
}

func (e *aesEncryptor) DecryptStream(cipherText io.Reader, output io.Writer) error {
	iv := make([]byte, aes.BlockSize)
	_, err := io.ReadFull(cipherText, iv)
	if err != nil {
		return fmt.Errorf("unable to read iv from cipherText, error: %w", err)
	}
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return fmt.Errorf("aes.NewCipher() error: %w", err)
	}
	dec := e.newDecryptor(block, iv)
	buf := make([]byte, bufferSize)
	return pipeStream(cipherText, "cipherText", output, "plainText", buf, func(b []byte) {
		dec.XORKeyStream(b, b)
	})
}

func (e *aesEncryptor) Overhead() int {
	return aes.BlockSize
}

func newAESCFB(passPhrase []byte, keyLength int) (StreamEncryptor, error) {
	deriveLength := keyLength
	if deriveLength < minDerivedKeyLength {
		deriveLength = minDerivedKeyLength
	}
	key, err := deriveKey(passPhrase, []byte(purposeEncryptionKey), deriveLength)
	if err != nil {
		return nil, fmt.Errorf("deriveKey() error: %w", err)
	}
	ae := &aesEncryptor{
		key:          key[:keyLength],
		newEncryptor: cipher.NewCFBEncrypter,
		newDecryptor: cipher.NewCFBDecrypter,
	}
	return ae, nil
}

func NewAES128CFB(passPhrase []byte) (StreamEncryptor, error) {
	return newAESCFB(passPhrase, 16)
}

func NewAES192CFB(passPhrase []byte) (StreamEncryptor, error) {
	return newAESCFB(passPhrase, 24)
}

func NewAES256CFB(passPhrase []byte) (StreamEncryptor, error) {
	return newAESCFB(passPhrase, 32)
}

func init() {
	Register("AES-128-CFB", "AES-128 with CFB mode", NewAES128CFB)
	Register("AES-192-CFB", "AES-192 with CFB mode", NewAES192CFB)
	Register("AES-256-CFB", "AES-256 with CFB mode", NewAES256CFB)
}
