package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/common/tplx"
)

func (c *CLI) configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config",
		Aliases: []string{"cfg"},
		Short:   "Manages configuration",
	}
	cmd.AddCommand(c.configInitCmd())
	cmd.AddCommand(c.configValueCmd())
	cmd.AddCommand(c.configValuesCmd())
	cmd.AddCommand(c.configExportCmd())
	return cmd
}

func (c *CLI) configInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "initialize",
		Aliases: []string{"init"},
		Short:   "Initialize configuration",
		Run: func(cmd *cobra.Command, args []string) {
			changed, err := c.config.InitializeWithChanged()
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("path", cfg.File())
			if changed {
				c.Changed("config initialized")
			} else {
				c.Ok("config already initialized")
			}
		},
	}
	return cmd
}

func (c *CLI) configValuesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "values",
		Aliases: []string{"get-all"},
		Short:   "Read all configuration values",
		Run: func(cmd *cobra.Command, args []string) {
			c.SetOutput("values", c.config.Values())
			c.Ok("config values read")
		},
	}
	return cmd
}

func (c *CLI) configValueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "value",
		Short:   "Read configuration value",
		Aliases: []string{"get"},
		Run: func(cmd *cobra.Command, args []string) {
			key, _ := cmd.Flags().GetString("key")
			template, _ := cmd.Flags().GetString("template")
			if key == "" && template == "" {
				c.Fail("flag 'key' or 'template' need to be specified")
				return
			}
			var (
				value string
				err   error
			)
			if key != "" {
				value, err = tplx.RenderKey(key, c.config.Values())
				if err != nil {
					c.Error(fmt.Errorf("cannot read config value using key '%s': %w", key, err))
					return
				}
			} else {
				value, err = tplx.RenderString(template, c.config.Values())
				if err != nil {
					c.Error(fmt.Errorf("cannot read config value using template '%s': %w", template, err))
					return
				}
			}
			c.SetOutput("value", value)
			c.Ok("config value read")
		},
	}
	cmd.Flags().StringP("key", "k", "", "Value key")
	cmd.Flags().StringP("template", "t", "", "Value template")
	cmd.MarkFlagsMutuallyExclusive("key", "template")
	return cmd
}

func (c *CLI) configExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "export",
		Aliases: []string{"save"},
		Short:   "Export configuration values to file",
		Run: func(cmd *cobra.Command, args []string) {
			file, _ := cmd.Flags().GetString("file")
			if err := c.config.Export(file); err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("file", file)
			c.Ok("config values exported")
		},
	}
	cmd.Flags().StringP("file", "f", "aem.env", "File path to export")
	cmd.MarkFlagRequired("file")
	return cmd
}
