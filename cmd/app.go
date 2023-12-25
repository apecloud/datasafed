package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/apecloud/datasafed/pkg/config"
	"github.com/apecloud/datasafed/pkg/logging"
	"github.com/apecloud/datasafed/pkg/storage"
	"github.com/apecloud/datasafed/pkg/storage/kopia"
	"github.com/apecloud/datasafed/pkg/storage/rclone"
)

const (
	backendBasePathEnv   = "DATASAFED_BACKEND_BASE_PATH"
	kopiaRepoRootEnv     = "DATASAFED_KOPIA_REPO_ROOT"
	kopiaPasswordEnv     = "DATASAFED_KOPIA_PASSWORD"
	kopiaDisableCacheEnv = "DATASAFED_KOPIA_DISABLE_CACHE"
	kopiaMaintenanceEnv  = "DATASAFED_KOPIA_MAINTENANCE"
	kopiaSafetyEnv       = "DATASAFED_KOPIA_SAFETY"
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

	basePath := strings.TrimSpace(os.Getenv(backendBasePathEnv))
	storageConf := config.GetGlobal().GetAll(config.StorageSection)

	if kopiaRoot := strings.TrimSpace(os.Getenv(kopiaRepoRootEnv)); kopiaRoot != "" {
		return initKopiaStorage(storageConf, basePath, kopiaRoot)
	} else {
		st, err := createStorage(storageConf, basePath)
		if err != nil {
			return err
		}
		globalStorage = st
		return nil
	}
}

func initKopiaStorage(storageConf map[string]string, basePath, kopiaRoot string) error {
	underlying, err := createStorage(storageConf, "")
	if err != nil {
		return err
	}
	kopia.SetUnderlyingStorage(underlying)
	storageConf[kopia.RepoRootKey] = kopiaRoot
	storageConf[kopia.PasswordKey] = strings.TrimSpace(os.Getenv(kopiaPasswordEnv))
	storageConf[kopia.DisableCacheKey] = strings.TrimSpace(os.Getenv(kopiaDisableCacheEnv))
	st, err := kopia.New(appCtx, storageConf, basePath)
	if err != nil {
		return err
	}
	globalStorage = st

	maintenance := os.Getenv(kopiaMaintenanceEnv)
	if ok, _ := strconv.ParseBool(maintenance); ok {
		onFinish(func() {
			err := kopia.RunMaintenance(appCtx, globalStorage, os.Getenv(kopiaSafetyEnv))
			if err != nil {
				fmt.Fprintf(os.Stderr, "RunMaintenance() failed, err: %v\n", err)
			}
		})
	}
	return nil
}

func createStorage(conf map[string]string, basePath string) (storage.Storage, error) {
	cloneConf := make(map[string]string, len(conf))
	for k, v := range conf {
		cloneConf[k] = v
	}
	return rclone.New(appCtx, cloneConf, basePath)
}

func onFinish(fn func()) {
	onFinishFuncs = append(onFinishFuncs, fn)
}

func Execute() {
	exitIfError(rootCmd.Execute())
}
