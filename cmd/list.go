package cmd

import (
	"bufio"
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/apecloud/datasafed/pkg/storage"
	"github.com/apecloud/datasafed/pkg/util"
)

var (
	validSorts         = []string{"path", "size", "mtime"}
	validOutputFormats = []string{"short", "long", "json"}
)

type listOptions struct {
	dirsOnly    bool
	filesOnly   bool
	recursive   bool
	maxDepth    int
	sort        string
	reverse     bool
	newer       int64
	older       int64
	namePattern string
	format      string
}

func init() {
	opts := &listOptions{}
	cmd := &cobra.Command{
		Use:   "list [-d|-f] [-r] [--max-depth depth] [-s sortBy] [--reverse] [--newer-than time] [--older-than time] [--name pattern] [-o outputFormat] rpath",
		Short: "List contents of a remote directory or file.",
		Example: strings.TrimSpace(`
# List the root directory
datasafed list /

# List one file and extract its size
datasafed list somefile.txt -o long | awk '{print $2}'

# List all files under the directory (ends with '/')
datasafed list -r -f /some/dir/

# List files modified within 1 hour and sort the result by size
datasafed list -r -f -s size --newer-than $(( $(date +%s) - 3600 )) /some/dir/

# List files with the name pattern
datasafed list --name "*.txt" /some/dir/
`),
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			doList(opts, cmd, args)
		},
	}
	pflags := cmd.PersistentFlags()
	pflags.BoolVarP(&opts.dirsOnly, "dirs-only", "d", false, "list directories only")
	pflags.BoolVarP(&opts.filesOnly, "files-only", "f", false, "list files only")
	pflags.BoolVarP(&opts.recursive, "recursive", "r", false, "list recursively")
	pflags.IntVar(&opts.maxDepth, "max-depth", 0, "max depth when listing recursively")
	pflags.VarP(util.NewEnumVar(validSorts, &opts.sort), "sort", "s",
		fmt.Sprintf("sort by which field, choices: %q, this option conflicts with --recursive", validSorts))
	pflags.BoolVar(&opts.reverse, "reverse", false, "reverse order")
	pflags.Int64Var(&opts.newer, "newer-than", 0,
		"list only entries whose last modification time is newer than the specified unix timestamp (exclusive)")
	pflags.Int64Var(&opts.older, "older-than", 0,
		"list only entries whose last modification time is older than the specified unix timestamp (exclusive)")
	pflags.StringVar(&opts.namePattern, "name", "",
		"list only entries whose name matches the specified pattern (https://pkg.go.dev/path/filepath#Match)")
	pflags.VarP(util.NewEnumVar(validOutputFormats, &opts.format).Default("short"), "output-format", "o",
		fmt.Sprintf("output format, choices: %q", validOutputFormats))

	cmd.MarkFlagsMutuallyExclusive("dirs-only", "files-only")
	cmd.MarkFlagsMutuallyExclusive("recursive", "sort")

	rootCmd.AddCommand(cmd)
}

func doList(opts *listOptions, cmd *cobra.Command, args []string) {
	bufStdout := bufio.NewWriterSize(os.Stdout, 8*1024)
	filter := getFilterFn(opts)
	printer := getPrinter(opts, bufStdout)
	var cb func(storage.DirEntry) error
	var entries []storage.DirEntry
	if opts.recursive {
		cb = func(entry storage.DirEntry) error {
			if filter(entry) {
				printer.printItem(entry)
			}
			return nil
		}
	} else {
		cb = func(entry storage.DirEntry) error {
			if filter(entry) {
				entries = append(entries, entry)
			}
			return nil
		}
	}

	if opts.recursive {
		printer.printHeader()
	}
	rpath := args[0]
	lopts := &storage.ListOptions{
		DirsOnly:  opts.dirsOnly,
		FilesOnly: opts.filesOnly,
		Recursive: opts.recursive,
		MaxDepth:  opts.maxDepth,
	}
	err := globalStorage.List(appCtx, rpath, lopts, cb)
	exitIfError(err)
	if opts.recursive {
		printer.printFooter()
	}

	if !opts.recursive {
		sortEntries(entries, opts)
		printer.printHeader()
		for _, entry := range entries {
			if filter(entry) {
				printer.printItem(entry)
			}
		}
		printer.printFooter()
	}

	bufStdout.Flush()
}

func toPath(entry storage.DirEntry) string {
	path := entry.Path()
	if entry.IsDir() && !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}

func getFilterFn(opts *listOptions) func(entry storage.DirEntry) bool {
	return func(entry storage.DirEntry) bool {
		matchPattern := func(entry storage.DirEntry) bool { return true }
		if opts.namePattern != "" {
			matchPattern = func(entry storage.DirEntry) bool {
				// ignore errors
				matched, _ := filepath.Match(opts.namePattern, entry.Name())
				return matched
			}
		}
		mt := entry.MTime().Unix()
		if opts.newer > 0 && opts.newer >= mt {
			return false
		}
		if opts.older > 0 && opts.older <= mt {
			return false
		}
		if !matchPattern(entry) {
			return false
		}
		return true
	}
}

func printJson(entry storage.DirEntry, enc *json.Encoder) {
	type jsonEntry struct {
		Path     string `json:"path"`
		Size     int64  `json:"size"`
		MTime    int64  `json:"mtime"`
		MTimeStr string `json:"mtime_str"`
		IsDir    bool   `json:"is_dir"`
	}
	_ = enc.Encode(jsonEntry{
		Path:     entry.Path(),
		Size:     entry.Size(),
		MTime:    entry.MTime().Unix(),
		MTimeStr: entry.MTime().Format(time.RFC3339),
		IsDir:    entry.IsDir(),
	})
}

func getPrinter(opts *listOptions, out io.Writer) *printer {
	switch opts.format {
	case "short":
		return &printer{
			printHeader: func() {},
			printItem: func(entry storage.DirEntry) {
				fmt.Fprintln(out, toPath(entry))
			},
			printFooter: func() {},
		}
	case "long":
		return &printer{
			printHeader: func() {},
			printItem: func(entry storage.DirEntry) {
				fmt.Fprintf(out, "%s\t%d\t%s\n", entry.MTime().Format(time.RFC3339), entry.Size(), toPath(entry))
			},
			printFooter: func() {},
		}
	case "json":
		enc := json.NewEncoder(out)
		enc.SetIndent("  ", "  ")
		first := true
		return &printer{
			printHeader: func() {
				fmt.Fprintf(out, "[")
			},
			printItem: func(entry storage.DirEntry) {
				if first {
					fmt.Fprintf(out, "\n")
				} else if !first {
					fmt.Fprintf(out, "  ,\n")
				}
				first = false
				fmt.Fprintf(out, "  ")
				printJson(entry, enc)
			},
			printFooter: func() {
				fmt.Fprintf(out, "]\n")
			},
		}
	}
	panic(fmt.Sprintf("unsupported format %s", opts.format))
}

func sortEntries(entries []storage.DirEntry, opts *listOptions) {
	if opts.sort == "" {
		opts.sort = "path"
	}
	var getter func(storage.DirEntry) any
	var compare func(any, any) int
	switch opts.sort {
	case "path":
		getter = func(entry storage.DirEntry) any { return entry.Path() }
		compare = func(a, b any) int { return cmp.Compare(a.(string), b.(string)) }
	case "size":
		getter = func(entry storage.DirEntry) any { return entry.Size() }
		compare = func(a, b any) int { return cmp.Compare(a.(int64), b.(int64)) }
	case "mtime":
		getter = func(entry storage.DirEntry) any { return entry.MTime() }
		compare = func(a, b any) int {
			ta, tb := a.(time.Time), b.(time.Time)
			if ta.Equal(tb) {
				return 0
			} else if ta.Before(tb) {
				return -1
			}
			return 1
		}
	}
	slices.SortFunc(entries, func(a, b storage.DirEntry) int {
		oa, ob := getter(a), getter(b)
		cmpVal := compare(oa, ob)
		if cmpVal == 0 {
			return cmp.Compare(a.Path(), b.Path())
		}
		if opts.reverse {
			return -cmpVal
		}
		return cmpVal
	})
}

type printer struct {
	printHeader func()
	printItem   func(entry storage.DirEntry)
	printFooter func()
}
