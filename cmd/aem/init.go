package main

import "github.com/spf13/cobra"

func (c *CLI) initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "initialize",
		Aliases: []string{"init"},
		Short:   "Initializes configuration and dependencies",
		Run: func(cmd *cobra.Command, args []string) {
			if !c.config.IsInitialized() {
				if err := c.config.Initialize(); err != nil {
					c.Error(err)
					return
				}
			}
			if err := c.aem.InstanceManager().LocalOpts.OakRun.Prepare(); err != nil {
				c.Error(err)
				return
			}
			c.Ok("initialized properly")

		},
	}
	return cmd
}
