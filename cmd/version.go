package cmd

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/apecloud/repocli/version"
)

type versionOptions struct {
	verbose bool
}

func init() {
	opts := &versionOptions{}
	cmd := &cobra.Command{
		Use:   "version [--verbose]",
		Short: "Show version of repocli.",
		Example: strings.TrimSpace(`
# Show version
repocli version
`),
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			doVersion(opts, cmd, args)
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			doNotInitStorage = true
			return nil
		},
	}
	cmd.PersistentFlags().BoolVarP(&opts.verbose, "verbose", "v", false, "show verbose version information")
	rootCmd.AddCommand(cmd)
}

func doVersion(opts *versionOptions, cmd *cobra.Command, args []string) {
	fmt.Printf("repocli: %s\n", version.Version)
	if opts.verbose {
		fmt.Printf("  BuildDate: %s\n", version.BuildDate)
		fmt.Printf("  GitCommit: %s\n", version.GitCommit)
		fmt.Printf("  GitTag: %s\n", version.GitVersion)
		fmt.Printf("  GoVersion: %s\n", runtime.Version())
		fmt.Printf("  Compiler: %s\n", runtime.Compiler)
		fmt.Printf("  Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	}
}
