package cmd

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "mkdir rpath",
		Short: "Create an empty remote directory.",
		Example: strings.TrimSpace(`
# Create an empty directory
datasafed mkdir some/dir
`),
		Args: cobra.ExactArgs(1),
		Run:  doMkdir,
	}
	rootCmd.AddCommand(cmd)
}

func doMkdir(cmd *cobra.Command, args []string) {
	err := globalStorage.Mkdir(context.Background(), args[0])
	exitIfError(err)
}
