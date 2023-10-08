package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	StorageSection = "storage"

	localBackendPathEnv = "DATASAFED_LOCAL_BACKEND_PATH"
)

var (
	global *Config
)

func InitGlobal(configFile string) error {
	// ignore the config file if a local backend is specified by env
	localBackendPath, err := checkToUseLocalBackend()
	if err != nil {
		return err
	}
	if localBackendPath != "" {
		global, err = NewStaticConfig(map[string]map[string]string{
			StorageSection: {
				"type": "local",
				"root": localBackendPath,
			},
		})
	} else {
		global, err = NewConfig(configFile)
	}
	return err
}

func checkToUseLocalBackend() (string, error) {
	localBackendPath := os.Getenv(localBackendPathEnv)
	if localBackendPath == "" {
		return "", nil
	}
	localBackendPath, err := filepath.Abs(localBackendPath)
	if err != nil {
		return "", err
	}
	if st, err := os.Stat(localBackendPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("the path \"%s\" specified by %s does not exist",
				localBackendPath, localBackendPathEnv)
		} else {
			return "", err
		}
	} else {
		if !st.IsDir() {
			return "", fmt.Errorf("the path \"%s\" specified by %s is not a directory",
				localBackendPath, localBackendPathEnv)
		}
	}
	return localBackendPath, nil
}

func GetGlobal() *Config {
	return global
}
