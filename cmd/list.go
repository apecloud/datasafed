package cmd

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/apecloud/repocli/pkg/storage"
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
		Short: "List contents of a directory",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			doList(opts, cmd, args)
		},
	}
	pflags := cmd.PersistentFlags()
	pflags.BoolVarP(&opts.dirsOnly, "dirs-only", "d", false, "list directories only")
	pflags.BoolVarP(&opts.filesOnly, "files-only", "f", false, "list files only")
	pflags.BoolVarP(&opts.recursive, "recursive", "r", false, "list recursively")
	pflags.IntVar(&opts.maxDepth, "max-depth", 0, "max depth when listing recursively")
	pflags.StringVarP(&opts.sort, "sort", "s", "path",
		fmt.Sprintf("sort by which field, choices: %q", validSorts))
	pflags.BoolVar(&opts.reverse, "reverse", false, "reverse order")
	pflags.Int64Var(&opts.newer, "newer-than", 0,
		"list only entries whose last modification time is newer than the specified unix timestamp (exclusive)")
	pflags.Int64Var(&opts.older, "older-than", 0,
		"list only entries whose last modification time is older than the specified unix timestamp (exclusive)")
	pflags.StringVar(&opts.namePattern, "name", "",
		"list only entries whose name matches the specified pattern (https://pkg.go.dev/path/filepath#Match)")
	pflags.StringVarP(&opts.format, "output-format", "o", "short",
		fmt.Sprintf("output format, choices: %q", validOutputFormats))

	cmd.MarkFlagsMutuallyExclusive("dirs-only", "files-only")

	rootCmd.AddCommand(cmd)
}

func doList(opts *listOptions, cmd *cobra.Command, args []string) {
	if !slices.Contains(validSorts, opts.sort) {
		exitIfError(fmt.Errorf("invalid sort: %q", opts.sort))
	}
	if !slices.Contains(validOutputFormats, opts.format) {
		exitIfError(fmt.Errorf("invalid output format: %q", opts.format))
	}
	rpath := args[0]
	lopts := &storage.ListOptions{
		DirsOnly:  opts.dirsOnly,
		FilesOnly: opts.filesOnly,
		Recursive: opts.recursive,
		MaxDepth:  opts.maxDepth,
	}
	entries, err := globalStorage.List(context.Background(), rpath, lopts)
	exitIfError(err)

	entries = filterEntries(entries, opts)
	sortEntries(entries, opts)
	switch opts.format {
	case "short":
		for _, entry := range entries {
			fmt.Println(toPath(entry))
		}
	case "long":
		for _, entry := range entries {
			fmt.Printf("%s\t%d\t%s\n", entry.MTime().Format(time.RFC3339), entry.Size(), toPath(entry))
		}
	case "json":
		if len(entries) == 0 {
			fmt.Print("[]\n")
		} else {
			fmt.Print("[\n")
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("  ", "  ")
			first := true
			for _, entry := range entries {
				if !first {
					fmt.Print("  ,\n")
				}
				first = false
				fmt.Print("  ")
				printJson(entry, enc)
			}
			fmt.Print("]\n")
		}
	}
}

func toPath(entry storage.DirEntry) string {
	path := entry.Path()
	if entry.IsDir() && !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
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

func filterEntries(entries []storage.DirEntry, opts *listOptions) []storage.DirEntry {
	matchPattern := func(entry storage.DirEntry) bool { return true }
	if opts.namePattern != "" {
		matchPattern = func(entry storage.DirEntry) bool {
			// ignore errors
			matched, _ := filepath.Match(opts.namePattern, entry.Name())
			return matched
		}
	}
	var filtered []storage.DirEntry
	for _, entry := range entries {
		mt := entry.MTime().Unix()
		if opts.newer > 0 && opts.newer >= mt {
			continue
		}
		if opts.older > 0 && opts.older <= mt {
			continue
		}
		if !matchPattern(entry) {
			continue
		}
		filtered = append(filtered, entry)
	}
	return filtered
}

func sortEntries(entries []storage.DirEntry, opts *listOptions) {
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
