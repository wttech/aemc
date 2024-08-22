package main

import (
	"github.com/spf13/cobra"
)

func (c *CLI) oakCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oak",
		Short: "Manages OAK repository",
	}
	cmd.AddCommand(c.oakIndexCmd())
	return cmd
}

func (c *CLI) oakIndexCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "index",
		Aliases: []string{"idx"},
		Short:   "Manage OAK indexes",
	}
	cmd.AddCommand(c.oakIndexListCmd())
	return cmd
}

func (c *CLI) oakIndexListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List OAK indexes",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}
			indexes, err := instance.OAK().IndexManager().List()
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("indexes", indexes)
			c.Ok("indexes listed")
		},
	}
	return cmd
}
