package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/apecloud/repocli/pkg/config"
	"github.com/apecloud/repocli/pkg/storage"
	"github.com/apecloud/repocli/pkg/storage/rclone"
)

const (
	rootKey            = "root"
	backendBasePathEnv = "REPOCLI_BACKEND_BASE_PATH"
)

var (
	rootCmd = &cobra.Command{Use: "repocli"}

	configFile       string
	doNotInitStorage bool
	globalStorage    storage.Storage
)

func init() {
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if doNotInitStorage {
			return nil
		}
		return initStorage()
	}
	rootCmd.PersistentFlags().StringVarP(&configFile, "conf", "c",
		"/etc/repocli/repocli.conf", "config file")
}

func exitIfError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func initStorage() error {
	if err := config.InitGlobal(configFile); err != nil {
		return err
	}

	storageConf := config.GetGlobal().GetAll(config.StorageSection)
	adjustRoot(storageConf)
	var err error
	globalStorage, err = rclone.New(storageConf)
	if err != nil {
		return err
	}
	return nil
}

func adjustRoot(storageConf map[string]string) {
	basePath := os.Getenv(backendBasePathEnv)
	if basePath == "" {
		return
	}
	basePath = filepath.Clean(basePath)
	if strings.HasPrefix(basePath, "..") {
		exitIfError(fmt.Errorf("invalid base path %q from env %s",
			os.Getenv(backendBasePathEnv), backendBasePathEnv))
	}
	if basePath == "." {
		basePath = ""
	} else {
		basePath = strings.TrimPrefix(basePath, "/")
		basePath = strings.TrimPrefix(basePath, "./")
	}
	root := storageConf[rootKey]
	if strings.HasSuffix(root, "/") {
		root = root + basePath
	} else {
		root = root + "/" + basePath
	}
	storageConf[rootKey] = root
}

func Execute() {
	exitIfError(rootCmd.Execute())
}
