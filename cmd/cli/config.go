package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

func (c *CLI) configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config",
		Aliases: []string{"cfg"},
		Short:   "Manages configuration",
	}

	cmd.AddCommand(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "print"},
		Short:   "Print effective configuration",
		Run: func(cmd *cobra.Command, args []string) {
			c.SetOutput("values", c.config.Values())
			c.Ok("config values printed")
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Initialize configuration",
		Run: func(cmd *cobra.Command, args []string) {
			err := c.config.Init()
			if err != nil {
				c.Fail(fmt.Sprintf("cannot initialize config: %s", err))
				return
			}
			c.SetOutput("path", c.config.File())
			c.Ok("config initialized properly")
		},
	})

	return cmd
}
