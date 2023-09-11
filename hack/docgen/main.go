package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/apecloud/repocli/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

//go:generate go run .

var (
	outDir = flag.String("out", "../../docs", "the output dir of the generated files")
)

func disableAutoGenTag(cmd *cobra.Command) {
	cmd.DisableAutoGenTag = true
	for _, c := range cmd.Commands() {
		disableAutoGenTag(c)
	}
}

func main() {
	flag.Parse()
	rootCmd := cmd.RootCommand()
	disableAutoGenTag(rootCmd)
	err := doc.GenMarkdownTree(rootCmd, *outDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
