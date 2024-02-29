package app

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/apecloud/datasafed/pkg/config"
	"github.com/apecloud/datasafed/pkg/encryption"
	"github.com/apecloud/datasafed/pkg/storage"
	"github.com/apecloud/datasafed/pkg/storage/encrypted"
	"github.com/apecloud/datasafed/pkg/storage/kopia"
	"github.com/apecloud/datasafed/pkg/storage/rclone"
)

const (
	backendBasePathEnv   = "DATASAFED_BACKEND_BASE_PATH"
	encryptionAlgorithm  = "DATASAFED_ENCRYPTION_ALGORITHM"
	encryptionPassPhrase = "DATASAFED_ENCRYPTION_PASS_PHRASE"
	kopiaRepoRootEnv     = "DATASAFED_KOPIA_REPO_ROOT"
	kopiaPasswordEnv     = "DATASAFED_KOPIA_PASSWORD"
	kopiaDisableCacheEnv = "DATASAFED_KOPIA_DISABLE_CACHE"
	kopiaMaintenanceEnv  = "DATASAFED_KOPIA_MAINTENANCE"
	kopiaSafetyEnv       = "DATASAFED_KOPIA_SAFETY"
)

var globalStorage storage.Storage

func InitGlobalStorage(ctx context.Context, configFile string) error {
	if globalStorage != nil {
		return fmt.Errorf("already inited")
	}
	if err := config.InitGlobal(configFile); err != nil {
		return err
	}

	basePath := strings.TrimSpace(os.Getenv(backendBasePathEnv))
	storageConf := config.GetGlobal().GetAll(config.StorageSection)

	if kopiaRoot := strings.TrimSpace(os.Getenv(kopiaRepoRootEnv)); kopiaRoot != "" {
		err := initKopiaStorage(ctx, storageConf, basePath, kopiaRoot)
		if err != nil {
			return err
		}
	} else {
		st, err := createStorage(ctx, storageConf, basePath)
		if err != nil {
			return err
		}
		globalStorage = st
	}

	// wrap with encryptedStorage
	encAlgo := os.Getenv(encryptionAlgorithm)
	if encAlgo != "" {
		encPass := os.Getenv(encryptionPassPhrase)
		if encPass == "" {
			return fmt.Errorf("encryption pass phrase should not be empty")
		}
		enc, err := encryption.CreateEncryptor(encAlgo, []byte(encPass))
		if err != nil {
			return err
		}
		encSt, err := encrypted.New(ctx, enc, globalStorage)
		if err != nil {
			return err
		}
		globalStorage = encSt
	}

	return nil
}

func GetGlobalStorage() (storage.Storage, error) {
	if globalStorage == nil {
		return nil, fmt.Errorf("not inited, call InitGlobalStorage() first")
	}
	return globalStorage, nil
}

func initKopiaStorage(ctx context.Context, storageConf map[string]string, basePath, kopiaRoot string) error {
	underlying, err := createStorage(ctx, storageConf, "")
	if err != nil {
		return err
	}
	kopia.SetUnderlyingStorage(underlying)
	storageConf[kopia.RepoRootKey] = kopiaRoot
	storageConf[kopia.PasswordKey] = strings.TrimSpace(os.Getenv(kopiaPasswordEnv))
	storageConf[kopia.DisableCacheKey] = strings.TrimSpace(os.Getenv(kopiaDisableCacheEnv))
	st, err := kopia.New(ctx, storageConf, basePath)
	if err != nil {
		return err
	}
	globalStorage = st

	maintenance := os.Getenv(kopiaMaintenanceEnv)
	if ok, _ := strconv.ParseBool(maintenance); ok {
		OnFinalize(func() {
			err := kopia.RunMaintenance(ctx, globalStorage, os.Getenv(kopiaSafetyEnv))
			if err != nil {
				fmt.Fprintf(os.Stderr, "RunMaintenance() failed, err: %v\n", err)
			}
		})
	}
	return nil
}

func createStorage(ctx context.Context, conf map[string]string, basePath string) (storage.Storage, error) {
	cloneConf := make(map[string]string, len(conf))
	for k, v := range conf {
		cloneConf[k] = v
	}
	return rclone.New(ctx, cloneConf, basePath)
}
