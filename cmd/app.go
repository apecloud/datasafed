package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/apecloud/datasafed/pkg/app"
	"github.com/apecloud/datasafed/pkg/logging"
	"github.com/apecloud/datasafed/pkg/storage"
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
)

func init() {
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		appCtx = logging.WithLogger(appCtx, logging.DefaultLoggerFactory)
		if !doNotInitStorage {
			if err := app.InitGlobalStorage(appCtx, configFile); err != nil {
				return err
			}
			var err error
			globalStorage, err = app.GetGlobalStorage()
			exitIfError(err)
		}
		return nil
	}
	rootCmd.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
		app.InvokeFinalizers()
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

func Execute() {
	exitIfError(rootCmd.Execute())
}
