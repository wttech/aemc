package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg"
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
	cmd.AddCommand(c.oakIndexReadCmd())
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

func (c *CLI) oakIndexReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "read",
		Short:   "Read OAK index details",
		Aliases: []string{"get", "find"},
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}
			bundle, err := oakIndexByFlags(cmd, *instance)
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("index", bundle)
			c.Ok("index read")
		},
	}
	oakIndexDefineFlags(cmd)
	return cmd
}

func oakIndexDefineFlags(cmd *cobra.Command) {
	cmd.Flags().String("name", "", "Name")
	_ = cmd.MarkFlagRequired("name")
}

func oakIndexByFlags(cmd *cobra.Command, i pkg.Instance) (*pkg.OAKIndex, error) {
	name, _ := cmd.Flags().GetString("name")
	if len(name) > 0 {
		index := i.OAK().IndexManager().New(name)
		return &index, nil
	}
	return nil, fmt.Errorf("flag 'name' is required")
}
