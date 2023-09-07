package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

type rmOptions struct {
	recursive bool
}

func init() {
	opts := &rmOptions{}
	cmd := &cobra.Command{
		Use:   "rm [-r] rpath",
		Short: "Remove a file, or all files in a directory",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			doRm(opts, cmd, args)
		},
	}
	cmd.PersistentFlags().BoolVarP(&opts.recursive, "recursive", "r", false, "remove recursively")
	rootCmd.AddCommand(cmd)
}

func doRm(opts *rmOptions, cmd *cobra.Command, args []string) {
	err := globalStorage.Remove(context.Background(), args[0], opts.recursive)
	exitIfError(err)
}
