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
	cmd.AddCommand(c.versionCmd())
	cmd.AddCommand(c.initCmd())
	cmd.AddCommand(c.configCmd())
	cmd.AddCommand(c.instanceCmd())
	cmd.AddCommand(c.osgiCmd())
	cmd.AddCommand(c.pkgCmd())
	cmd.AddCommand(c.repoCmd())
	cmd.AddCommand(c.replCmd())
	cmd.AddCommand(c.cryptoCmd())
	cmd.AddCommand(c.fileCmd())
	c.rootFlags(cmd)
	return cmd
}

func (c *CLI) rootFlags(cmd *cobra.Command) {
	// input/output
	cmd.PersistentFlags().StringVar(&(c.inputFormat),
		"input-format", c.inputFormat,
		"Controls input format ("+strings.Join(cfg.InputFormats(), "|")+")")
	cmd.PersistentFlags().StringVar(&(c.inputFile),
		"input-file", c.inputFile,
		"Provides input as file path")
	cmd.PersistentFlags().StringVar(&(c.inputString),
		"input-string", c.inputString,
		"Provides input as string")
	cmd.PersistentFlags().StringVar(&(c.outputFormat),
		"output-format", c.outputFormat,
		"Controls output format ("+strings.Join(cfg.OutputFormats(), "|")+")")
	cmd.PersistentFlags().StringVar(&(c.outputLogFile),
		"output-log-file", c.outputLogFile,
		"Controls output file path")
	cmd.PersistentFlags().StringVar(&(c.outputLogMode),
		"output-log-mode", c.outputLogMode,
		"Controls where outputs and logs should be written to when format is \"text\""+(strings.Join(cfg.OutputLogModes(), "|")+")"))
	cmd.PersistentFlags().StringVar(&(c.outputValue),
		"output-value", c.outputValue,
		"Limits output to single variable")

	cmd.PersistentFlags().StringVarP(&(c.aem.InstanceManager().AdHocURL),
		"instance-url", "U", c.aem.InstanceManager().AdHocURL,
		"Use only AEM instance at ad-hoc specified URL")
	cmd.PersistentFlags().StringVarP(&(c.aem.InstanceManager().FilterID),
		"instance-id", "I", c.aem.InstanceManager().FilterID,
		"Use only AEM instance configured with the exact ID")

	cmd.PersistentFlags().BoolVarP(&(c.aem.InstanceManager().FilterAuthors),
		"instance-author", "A", c.aem.InstanceManager().FilterAuthors,
		"Use only AEM author instance")
	cmd.PersistentFlags().BoolVarP(&(c.aem.InstanceManager().FilterPublishes),
		"instance-publish", "P", c.aem.InstanceManager().FilterPublishes,
		"Use only AEM publish instance")
	cmd.MarkFlagsMutuallyExclusive("instance-author", "instance-publish")

	cmd.PersistentFlags().StringVar(&(c.aem.InstanceManager().ProcessingMode),
		"instance-processing", c.aem.InstanceManager().ProcessingMode,
		"Controls processing mode for instances ("+(strings.Join(instance.ProcessingModes(), "|")+")"))
}
