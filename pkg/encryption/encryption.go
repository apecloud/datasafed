// Package encryption manages content encryption algorithms.
package encryption

import (
	"sort"
	"strings"

	"github.com/pkg/errors"
)

const (
	purposeEncryptionKey = "encryption"
)

// CreateEncryptor creates an StreamEncryptor for given parameters.
func CreateEncryptor(algorithm string, passPhrase []byte) (StreamEncryptor, error) {
	algorithm = strings.ToUpper(algorithm)
	e := encryptors[algorithm]
	if e == nil {
		return nil, errors.Errorf("unknown encryption algorithm: %v", algorithm)
	}

	return e.newEncryptor(passPhrase)
}

// EncryptorFactory creates new Encryptor for given parameters.
type EncryptorFactory func(passPhrase []byte) (StreamEncryptor, error)

// DefaultAlgorithm is the name of the default encryption algorithm.
const DefaultAlgorithm = "AES256-CFB"

// SupportedAlgorithms returns the names of the supported encryption methods.
func SupportedAlgorithms() []string {
	var result []string
	for k := range encryptors {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}

// Register registers new encryption algorithm.
func Register(name, description string, newEncryptor EncryptorFactory) {
	name = strings.ToUpper(name)
	encryptors[name] = &encryptorInfo{
		description,
		newEncryptor,
	}
}

type encryptorInfo struct {
	description  string
	newEncryptor EncryptorFactory
}

var encryptors = map[string]*encryptorInfo{}
