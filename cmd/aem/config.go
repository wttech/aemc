package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg/cfg"
)

func (c *CLI) configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config",
		Aliases: []string{"cfg"},
		Short:   "Manages configuration",
	}
	cmd.AddCommand(c.configListCmd())
	cmd.AddCommand(c.configInitCmd())
	return cmd
}

func (c *CLI) configListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "print"},
		Short:   "Print effective configuration",
		Run: func(cmd *cobra.Command, args []string) {
			c.SetOutput("values", c.config.Values())
			c.Ok("config values printed")
		},
	}
	return cmd
}

func (c *CLI) configInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration",
		Run: func(cmd *cobra.Command, args []string) {
			err := c.config.Init()
			if err != nil {
				c.Fail(fmt.Sprintf("cannot initialize config: %s", err))
				return
			}
			c.SetOutput("path", cfg.File())
			c.Ok("config initialized properly")
		},
	}
	return cmd
}
