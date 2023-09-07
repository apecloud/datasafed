package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

type statOptions struct {
	json bool
}

func init() {
	opts := &statOptions{}
	cmd := &cobra.Command{
		Use:   "stat [--json] rpath",
		Short: "Print statistics about a remote path",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			doStat(opts, cmd, args)
		},
	}
	cmd.PersistentFlags().BoolVar(&opts.json, "json", false, "output in json format")
	rootCmd.AddCommand(cmd)
}

func doStat(opts *statOptions, cmd *cobra.Command, args []string) {
	rpath := args[0]
	result, err := globalStorage.Stat(context.Background(), rpath)
	exitIfError(err)
	if !opts.json {
		fmt.Printf("TotalSize: %d\n", result.TotalSize)
		fmt.Printf("Entries: %d\n", result.Entries)
		fmt.Printf("Dirs: %d\n", result.Dirs)
		fmt.Printf("Files: %d\n", result.Files)
	} else {
		output := map[string]interface{}{
			"total_size": result.TotalSize,
			"entries":    result.Entries,
			"dirs":       result.Dirs,
			"files":      result.Files,
		}
		data, _ := json.Marshal(output)
		fmt.Printf("%s\n", string(data))
	}
}
