package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "pull rpath lpath",
		Short: "Pull remote file",
		Args:  cobra.ExactArgs(2),
		Run:   doPull,
	}
	rootCmd.AddCommand(cmd)
}

func doPull(cmd *cobra.Command, args []string) {
	rpath := args[0]
	lpath := args[1]
	var out io.Writer
	if lpath == "-" {
		out = os.Stdout
	} else {
		if lpath == "" || strings.HasSuffix(lpath, "/") {
			exitIfError(fmt.Errorf("invalid local path \"%s\"", lpath))
		}
		if !strings.HasPrefix(lpath, "/") {
			var err error
			lpath, err = filepath.Abs(lpath)
			exitIfError(err)
		}
		err := os.MkdirAll(filepath.Dir(lpath), 0755)
		exitIfError(err)
		f, err := os.OpenFile(lpath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		exitIfError(err)
		defer f.Close()
		out = f
	}
	err := globalStorage.Pull(context.Background(), rpath, out)
	if err != nil {
		err = fmt.Errorf("pull %q: %w", rpath, err)
	}
	exitIfError(err)
}
