package commands

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCommand = &cobra.Command{
		Use:          "shogun",
		Short:        "Command-Line for Shogun-CD.",
		SilenceUsage: true,
	}

	Server string
)

func Execute() error {
	return rootCommand.Execute()
}

func init() {
	rootCommand.PersistentFlags().StringVarP(&Server, "server", "s", "", "Shogun server URL (env=SHOGUN_SERVER)")
}

func getServer() string {
	if Server != "" {
		return Server
	}

	env := os.Getenv("SHOGUN_SERVER")
	if env != "" {
		return env
	}

	return "http://localhost:7007"
}
