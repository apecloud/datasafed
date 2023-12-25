package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "rmdir rpath",
		Short: "Remove an empty remote directory.",
		Long: "Remove an empty remote directory.\n" +
			"Some storage backends, such as S3, do not have the concept of a directory, " +
			"in which case the command will directly return success with no effect.",
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
	err := globalStorage.Rmdir(appCtx, args[0])
	exitIfError(err)
}
