package cmd

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "rmdir rpath",
		Short: "Remove an empty remote directory.",
		Example: strings.TrimSpace(`
# Remove an empty directory
datasafed rmdir some/dir
`),
		Args: cobra.ExactArgs(1),
		Run:  doRmdir,
	}
	rootCmd.AddCommand(cmd)
}

func doRmdir(cmd *cobra.Command, args []string) {
	err := globalStorage.Rmdir(context.Background(), args[0])
	exitIfError(err)
}
