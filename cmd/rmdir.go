package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "rmdir rpath",
		Short: "Remove an empty directory",
		Args:  cobra.ExactArgs(1),
		Run:   doRmdir,
	}
	rootCmd.AddCommand(cmd)
}

func doRmdir(cmd *cobra.Command, args []string) {
	err := globalStorage.Rmdir(context.Background(), args[0])
	exitIfError(err)
}
