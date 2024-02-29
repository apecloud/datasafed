package cmd

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/kopia/kopia/repo/compression"
	"github.com/spf13/cobra"

	"github.com/apecloud/datasafed/pkg/util"
)

// from https://github.com/kopia/kopia/blob/a934629c55f5a04a1496b58708bf44df1f7b6690/repo/compression/compressor.go#L13
const compressionHeaderSize = 4

var (
	validCompressionAlgorithms = getValidCompressionAlgorithms()
)

type pushOptions struct {
	compression string
}

func init() {
	opts := &pushOptions{}
	cmd := &cobra.Command{
		Use:   "push lpath rpath",
		Short: "Push file to remote",
		Long:  "The `lpath` parameter can be '-' to read from stdin.",
		Example: strings.TrimSpace(`
# Push a file to remote
datasafed push local/path/a.txt remote/path/a.txt

# Upload data from stdin
datasafed push - remote/path/somefile.txt
`),
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			doPush(opts, cmd, args)
		},
	}
	pflags := cmd.PersistentFlags()
	pflags.VarP(util.NewEnumVar(validCompressionAlgorithms, &opts.compression), "compress", "z",
		fmt.Sprintf("compress the file using the specified algorithm before sending it to remote, choices: %q", validCompressionAlgorithms))
	rootCmd.AddCommand(cmd)
}

func getValidCompressionAlgorithms() []string {
	var names []string
	for name := range compression.ByName {
		names = append(names, string(name))
	}
	sort.Strings(names)
	return names
}

func doPush(opts *pushOptions, cmd *cobra.Command, args []string) {
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
	if opts.compression != "" {
		c, ok := compression.ByName[compression.Name(opts.compression)]
		if !ok {
			exitIfError(fmt.Errorf("bug: compressor for %s is not found", opts.compression))
		}
		pr, pw := io.Pipe()
		go func(r io.Reader) {
			// discard compression header
			err := c.Compress(util.DiscardNWriter(pw, compressionHeaderSize), r)
			pw.CloseWithError(err)
		}(in)
		in = pr
	}
	err := globalStorage.Push(appCtx, in, rpath)
	if err != nil {
		err = fmt.Errorf("push to %q: %w", rpath, err)
	}
	exitIfError(err)
}
