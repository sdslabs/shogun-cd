package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "dev"

var versionCommand = &cobra.Command{
	Use:   "version",
	Short: "Print the Shogun CLI version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("shogun", version)
	},
}

func init() {
	rootCommand.AddCommand(versionCommand)
}
