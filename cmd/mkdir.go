package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use: "mkdir rpath",
		Short: "Create an empty remote directory." +
			"Some storage backends, such as S3, do not have the concept of a directory, " +
			"in which case the command will directly return success with no effect.",
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
	err := globalStorage.Mkdir(appCtx, args[0])
	exitIfError(err)
}
