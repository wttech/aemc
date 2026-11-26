package main

import (
	"fmt"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg/cfg"
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
	cmd.AddCommand(c.configMigrateCmd())
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
			values := c.aem.Config().Values()
			if !values.IsSet(key) {
				c.Error(fmt.Errorf("config value not found using key '%s'", key))
				return
			}
			c.SetOutput("value", fmt.Sprint(values.Get(key)))
			c.Ok("config value read")
		},
	}
	cmd.Flags().StringP("key", "k", "", "Value key")
	_ = cmd.MarkFlagRequired("key")
	return cmd
}

// configMigrateCmd is a temporary bridge for config files of AEM Compose 2.x; remove in the next major version.
func (c *CLI) configMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate config files from AEM Compose 2.x template syntax '[[ ]]' to variables '${}'",
		Run: func(cmd *cobra.Command, args []string) {
			migrations, err := cfg.MigrateLegacyFiles()
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("migrations", migrations)
			changed := lo.SomeBy(migrations, func(m cfg.LegacyFileMigration) bool { return m.Changed })
			unresolved := lo.SomeBy(migrations, func(m cfg.LegacyFileMigration) bool { return len(m.Unresolved) > 0 })
			if unresolved {
				c.Fail("config files migrated but some lines need a manual change, see 'unresolved'")
			} else if changed {
				c.Changed("config files migrated")
			} else {
				c.Ok("config files already migrated")
			}
		},
	}
	return cmd
}
