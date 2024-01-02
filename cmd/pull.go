package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kopia/kopia/repo/compression"
	"github.com/spf13/cobra"

	"github.com/apecloud/datasafed/pkg/util"
)

type pullOptions struct {
	decompression string
}

func init() {
	opts := &pullOptions{}
	cmd := &cobra.Command{
		Use:   "pull rpath lpath",
		Short: "Pull remote file",
		Long:  "The `lpath` parameter can be \"-\" to write to stdout.",
		Example: strings.TrimSpace(`
# Pull the file and save it to a local path
datasafed pull some/path/file.txt /tmp/file.txt

# Pull the file and print it to stdout
datasafed pull some/path/file.txt - | wc -l
`),
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			doPull(opts, cmd, args)
		},
	}
	pflags := cmd.PersistentFlags()
	pflags.VarP(util.NewEnumVar(validCompressionAlgorithms, &opts.decompression), "decompress", "d",
		fmt.Sprintf("decompress the pulled file using the specified algorithm, choices: %q", validCompressionAlgorithms))
	rootCmd.AddCommand(cmd)
}

func doPull(opts *pullOptions, cmd *cobra.Command, args []string) {
	rpath := args[0]
	lpath := args[1]
	var out io.Writer
	var flush func() error
	if lpath == "-" {
		out = os.Stdout
		flush = func() error { return nil }
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
		out = f
		flush = func() error { return f.Close() }
	}
	if opts.decompression != "" {
		c, ok := compression.ByName[compression.Name(opts.decompression)]
		if !ok {
			exitIfError(fmt.Errorf("bug: compressor for %s is not found", opts.decompression))
		}
		pr, pw := io.Pipe()
		ch := make(chan error, 1)
		go func(out io.Writer) {
			err := c.Decompress(out, pr, false)
			pr.CloseWithError(err)
			ch <- err
		}(out)
		out = pw
		originalFlush := flush
		flush = func() error {
			pw.Close() // reach EOF
			err := <-ch
			if err != nil {
				return err
			}
			return originalFlush()
		}
	}
	err := globalStorage.Pull(appCtx, rpath, out)
	if err != nil {
		err = fmt.Errorf("pull %q: %w", rpath, err)
	}
	if ferr := flush(); err == nil {
		err = ferr
	}
	exitIfError(err)
}
