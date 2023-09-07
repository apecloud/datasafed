package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "push lpath rpath",
		Short: "Push file to remote",
		Args:  cobra.ExactArgs(2),
		Run:   doPush,
	}
	rootCmd.AddCommand(cmd)
}

func doPush(cmd *cobra.Command, args []string) {
	lpath := args[0]
	rpath := args[1]
	var in io.Reader
	if lpath == "-" {
		in = os.Stdin
	} else {
		f, err := os.Open(lpath)
		exitIfError(err)
		defer f.Close()
		in = f
	}
	err := globalStorage.Push(context.Background(), in, rpath)
	if err != nil {
		err = fmt.Errorf("push to %q: %w", rpath, err)
	}
	exitIfError(err)
}
