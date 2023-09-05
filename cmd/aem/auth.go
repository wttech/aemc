package main

import "github.com/spf13/cobra"

func (c *CLI) authCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Auth management",
	}
	cmd.AddCommand(c.userCmd())
	return cmd
}
