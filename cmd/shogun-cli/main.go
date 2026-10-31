package main

import (
	"os"

	"github.com/kunalvirwal/shogun-cd/cli/commands"
)

func main() {
	err := commands.Execute()
	if err != nil {
		os.Exit(1)
	}
}
