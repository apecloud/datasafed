package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/apecloud/datasafed/pkg/config"
	"github.com/apecloud/datasafed/pkg/logging"
	"github.com/apecloud/datasafed/pkg/storage"
	"github.com/apecloud/datasafed/pkg/storage/rclone"
)

const (
	rootKey            = "root"
	backendBasePathEnv = "DATASAFED_BACKEND_BASE_PATH"
)

var (
	rootCmd = &cobra.Command{
		Use:           "datasafed",
		Short:         "`datasafed` is a command line tool for managing remote storages.",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	configFile       string
	doNotInitStorage bool
	globalStorage    storage.Storage
	appCtx           context.Context = context.Background()
	onFinishFuncs    []func()
)

func init() {
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		appCtx = logging.WithLogger(appCtx, logging.DefaultLoggerFactory)
		if !doNotInitStorage {
			if err := initStorage(); err != nil {
				return err
			}
		}
		return nil
	}
	rootCmd.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
		for _, fn := range onFinishFuncs {
			fn()
		}
		return nil
	}
	rootCmd.PersistentFlags().StringVarP(&configFile, "conf", "c",
		"/etc/datasafed/datasafed.conf", "config file")

	logging.Attach(rootCmd)
}

// RootCommand returns the root command.
// It is used by docgen.
func RootCommand() *cobra.Command {
	return rootCmd
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
