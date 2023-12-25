package kopia

import "github.com/apecloud/datasafed/pkg/storage"

var (
	underlying storage.Storage
)

func SetUnderlyingStorage(st storage.Storage) {
	underlying = st
}

func GetUnderlyingStorage() storage.Storage {
	return underlying
}
