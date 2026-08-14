package main

import (
	"techybat.org/go-vpn/cmd"
	"techybat.org/go-vpn/cmd/bot"
	"techybat.org/go-vpn/cmd/renew"
)

func main() {
	// Register subcommands
	cmd.RootCmd.AddCommand(bot.StartCmd)
	cmd.RootCmd.AddCommand(renew.RenewCmd)

	// Execute the CLI
	cmd.Execute()
}
