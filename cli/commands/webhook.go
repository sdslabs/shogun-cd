package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var webhookCommand = &cobra.Command{
	Use:   "hook",
	Short: "Webhook Commands",
	Long:  "Trigger, Create, Delete, List, Activate and Deactivate Webhooks",
}

var webhookTriggerCommand = &cobra.Command{
	Use:   "trigger",
	Short: "Trigger a webhook using slug",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		slug := args[0]
		url := getServer() + "/hook/" + slug
		fmt.Println("POST", url)
		return nil
	},
}
