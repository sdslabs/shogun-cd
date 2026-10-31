package commands

import (
	"github.com/spf13/cobra"
)

var (
	rootCommand = &cobra.Command{
		Use:   "shogun",
		Short: "Command-Line for Shogun-CD.",
	}
)

func Execute() error {
	return rootCommand.Execute()
}
