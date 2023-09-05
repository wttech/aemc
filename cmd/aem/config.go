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
		Short:   "Manages AEMC configuration",
	}
	cmd.AddCommand(c.configInitCmd())
	cmd.AddCommand(c.configExportCmd())
	cmd.AddCommand(c.configValueCmd())
	cmd.AddCommand(c.configValuesCmd())
	return cmd
}

func (c *CLI) configInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "initialize",
		Aliases: []string{"init"},
		Short:   "Initialize configuration",
		Run: func(cmd *cobra.Command, args []string) {
			changed, err := c.aem.Config().InitializeWithChanged()
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

func (c *CLI) configExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "export",
		Aliases: []string{"save"},
		Short:   "Exports current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			file, _ := cmd.Flags().GetString("file")
			if err := c.aem.Config().Export(file); err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("file", file)
			c.Changed("config exported")
		},
	}
	cmd.Flags().StringP("file", "f", "", "Target file path")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func (c *CLI) configValuesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "values",
		Aliases: []string{"get-all"},
		Short:   "Read all configuration values",
		Run: func(cmd *cobra.Command, args []string) {
			c.SetOutput("file", cfg.FileEffective())
			c.SetOutput("values", c.aem.Config().Values().AllSettings())
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
				value, err = tplx.RenderKey(key, c.aem.Config().Values().AllSettings())
				if err != nil {
					c.Error(fmt.Errorf("cannot read config value using key '%s': %w", key, err))
					return
				}
			} else {
				value, err = tplx.RenderString(template, c.aem.Config().Values().AllSettings())
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
