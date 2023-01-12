package main

import (
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/instance"
	"strings"
)

func (c *CLI) rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "aem",

		// needed to properly bind CLI flags with viper values from env and YML files
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			c.configure()
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			c.exit()
			return nil
		},
	}
	cmd.AddCommand(c.appCmd())
	cmd.AddCommand(c.versionCmd())
	cmd.AddCommand(c.configCmd())
	cmd.AddCommand(c.instanceCmd())
	cmd.AddCommand(c.osgiCmd())
	cmd.AddCommand(c.pkgCmd())
	cmd.AddCommand(c.repoCmd())
	cmd.AddCommand(c.replCmd())
	cmd.AddCommand(c.fileCmd())
	c.rootFlags(cmd)
	return cmd
}

func (c *CLI) rootFlags(cmd *cobra.Command) {
	// input/output
	cmd.PersistentFlags().StringVar(&(c.config.Values().Input.Format),
		"input-format", c.config.Values().Input.Format,
		"Controls input format ("+strings.Join(cfg.InputFormats(), "|")+")")
	cmd.PersistentFlags().StringVar(&(c.config.Values().Input.File),
		"input-file", c.config.Values().Input.File,
		"Provides input as file path")
	cmd.PersistentFlags().StringVar(&(c.config.Values().Input.String),
		"input-string", c.config.Values().Input.String,
		"Provides input as string")
	cmd.PersistentFlags().StringVar(&(c.config.Values().Output.Format),
		"output-format", c.config.Values().Output.Format,
		"Controls output format ("+strings.Join(cfg.OutputFormats(), "|")+")")
	cmd.PersistentFlags().StringVar(&(c.config.Values().Output.File),
		"output-file", c.config.Values().Output.File,
		"Controls output file path")
	cmd.PersistentFlags().StringVar(&(c.config.Values().Output.Value),
		"output-value", c.config.Values().Output.Value,
		"Limits output to single variable")

	// instance filtering
	cmd.PersistentFlags().StringVarP(&(c.config.Values().Instance.ConfigURL),
		"instance-url", "U", c.config.Values().Instance.ConfigURL,
		"Use only AEM instance at specified URL")
	cmd.PersistentFlags().StringVarP(&(c.config.Values().Instance.Filter.ID),
		"instance-id", "I", c.config.Values().Instance.Filter.ID,
		"Use only AEM instance with specified ID")
	cmd.PersistentFlags().BoolVarP(&(c.config.Values().Instance.Filter.Author),
		"instance-author", "A", c.config.Values().Instance.Filter.Author,
		"Use only AEM author instance")
	cmd.PersistentFlags().BoolVarP(&(c.config.Values().Instance.Filter.Publish),
		"instance-publish", "P", c.config.Values().Instance.Filter.Publish,
		"Use only AEM publish instance")
	cmd.PersistentFlags().StringVar(&(c.config.Values().Instance.ProcessingMode),
		"instance-processing", c.config.Values().Instance.ProcessingMode,
		"Controls processing mode for instances ("+(strings.Join(instance.ProcessingModes(), "|")+")"))
}
