package cmd

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
)

type rmOptions struct {
	recursive bool
}

func init() {
	opts := &rmOptions{}
	cmd := &cobra.Command{
		Use:   "rm [-r] rpath",
		Short: "Remove one remote file, or all files in a remote directory.",
		Example: strings.TrimSpace(`
# Remove a single file
repocli rm some/path/to/file.txt

# Recursively remove a directory
repocli rm -r some/path/to/dir
`),
		Args: cobra.ExactArgs(1),
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
