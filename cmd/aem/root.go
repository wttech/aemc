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
	cv := c.aem.Config().Values()

	// input/output
	cmd.PersistentFlags().StringVar(&(c.inputFormat),
		"input-format", c.inputFormat,
		"Controls input format ("+strings.Join(cfg.InputFormats(), "|")+")")
	_ = cv.BindPFlag("input.format", cmd.PersistentFlags().Lookup("input-format"))

	cmd.PersistentFlags().StringVar(&(c.inputFile),
		"input-file", c.inputFile,
		"Provides input as file path")
	_ = cv.BindPFlag("input.file", cmd.PersistentFlags().Lookup("input-file"))

	cmd.PersistentFlags().StringVar(&(c.inputString),
		"input-string", c.inputString,
		"Provides input as string")
	_ = cv.BindPFlag("input.string", cmd.PersistentFlags().Lookup("input-string"))

	cmd.PersistentFlags().StringVar(&(c.outputValue),
		"output-value", c.outputValue,
		"Limits output to single variable")
	_ = cv.BindPFlag("output.value", cmd.PersistentFlags().Lookup("output-value"))

	cmd.PersistentFlags().StringVar(&(c.outputFormat),
		"output-format", c.outputFormat,
		"Controls output format ("+strings.Join(cfg.OutputFormats(), "|")+")")
	_ = cv.BindPFlag("output.format", cmd.PersistentFlags().Lookup("output-format"))

	cmd.PersistentFlags().StringVar(&(c.outputLogFile),
		"output-log-file", c.outputLogFile,
		"Controls output file path")
	_ = cv.BindPFlag("output.log.file", cmd.PersistentFlags().Lookup("output-log-file"))

	cmd.PersistentFlags().StringVar(&(c.outputLogMode),
		"output-log-mode", c.outputLogMode,
		"Controls where outputs and logs should be written to when format is \"text\""+(strings.Join(cfg.OutputLogModes(), "|")+")"))
	_ = cv.BindPFlag("output.log.mode", cmd.PersistentFlags().Lookup("output-log-mode"))

	cmd.PersistentFlags().StringVarP(&(c.aem.InstanceManager().AdHocURL),
		"instance-url", "U", c.aem.InstanceManager().AdHocURL,
		"Use only AEM instance at ad-hoc specified URL")
	_ = cv.BindPFlag("instance.adhoc_url", cmd.PersistentFlags().Lookup("instance-url"))

	cmd.PersistentFlags().StringVarP(&(c.aem.InstanceManager().FilterID),
		"instance-id", "I", c.aem.InstanceManager().FilterID,
		"Use only AEM instance configured with the exact ID")
	_ = cv.BindPFlag("instance.filter.id", cmd.PersistentFlags().Lookup("instance-id"))

	cmd.PersistentFlags().BoolVarP(&(c.aem.InstanceManager().FilterAuthors),
		"instance-author", "A", c.aem.InstanceManager().FilterAuthors,
		"Use only AEM author instance")
	_ = cv.BindPFlag("instance.filter.authors", cmd.PersistentFlags().Lookup("instance-author"))

	cmd.PersistentFlags().BoolVarP(&(c.aem.InstanceManager().FilterPublishes),
		"instance-publish", "P", c.aem.InstanceManager().FilterPublishes,
		"Use only AEM publish instance")
	_ = cv.BindPFlag("instance.filter.publishes", cmd.PersistentFlags().Lookup("instance-publish"))

	cmd.MarkFlagsMutuallyExclusive("instance-author", "instance-publish")

	cmd.PersistentFlags().StringVar(&(c.aem.InstanceManager().ProcessingMode),
		"instance-processing", c.aem.InstanceManager().ProcessingMode,
		"Controls processing mode for instances ("+(strings.Join(instance.ProcessingModes(), "|")+")"))
	_ = cv.BindPFlag("instance.processing_mode", cmd.PersistentFlags().Lookup("instance-processing"))
}
