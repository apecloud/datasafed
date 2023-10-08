package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/apecloud/datasafed/pkg/config"
)

func init() {
	cmd := &cobra.Command{
		Use:   "getconf item",
		Short: "Get the value of the configuration item.",
		Long:  "The pattern of the `item` parameter is \"section.field\".",
		Example: strings.TrimSpace(`
# get the "type" field from the "storage" section
datasafed getconf storage.type

# get access_key_id (only available for S3 backend)
datasafed getconf storage.access_key_id
`),
		Args: cobra.ExactArgs(1),
		Run:  doGetconf,
	}
	rootCmd.AddCommand(cmd)
}

func doGetconf(cmd *cobra.Command, args []string) {
	item := args[0]
	parts := strings.SplitN(item, ".", 2)
	if len(parts) != 2 {
		exitIfError(fmt.Errorf("invalid config item name %q", item))
	}
	cfg := config.GetGlobal()
	value, exists := cfg.Get(parts[0], parts[1])
	if !exists {
		exitIfError(fmt.Errorf("config item %q not found", item))
	}
	fmt.Println(value)
}
