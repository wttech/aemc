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
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			c.onStart()
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			c.onEnd()
			return nil
		},
	}
	cmd.AddCommand(c.versionCmd())
	cmd.AddCommand(c.projectCmd())
	cmd.AddCommand(c.vendorCmd())
	cmd.AddCommand(c.configCmd())
	cmd.AddCommand(c.instanceCmd())
	cmd.AddCommand(c.osgiCmd())
	cmd.AddCommand(c.oakCmd())
	cmd.AddCommand(c.pkgCmd())
	cmd.AddCommand(c.repoCmd())
	cmd.AddCommand(c.replCmd())
	cmd.AddCommand(c.cryptoCmd())
	cmd.AddCommand(c.sslCmd())
	cmd.AddCommand(c.gtsCmd())
	cmd.AddCommand(c.fileCmd())
	cmd.AddCommand(c.authCmd())
	cmd.AddCommand(c.contentCmd())

	c.rootFlags(cmd)

	return cmd
}

func (c *CLI) rootFlags(cmd *cobra.Command) {
	cv := c.config.Values()

	cmd.PersistentFlags().String("input-format", cv.GetString("input.format"), "Controls input format ("+strings.Join(cfg.InputFormats(), "|")+")")
	_ = cv.BindPFlag("input.format", cmd.PersistentFlags().Lookup("input-format"))

	cmd.PersistentFlags().String("input-file", cv.GetString("input.file"), "Provides input as file path")
	_ = cv.BindPFlag("input.file", cmd.PersistentFlags().Lookup("input-file"))

	cmd.PersistentFlags().String("input-string", cv.GetString("input.string"), "Provides input as string")
	_ = cv.BindPFlag("input.string", cmd.PersistentFlags().Lookup("input-string"))

	cmd.PersistentFlags().StringP("output-value", "V", cv.GetString("output.value"),
		"Limits output to single variable")
	_ = cv.BindPFlag("output.value", cmd.PersistentFlags().Lookup("output-value"))

	cmd.PersistentFlags().String("output-format", cv.GetString("output.format"), "Controls output format ("+strings.Join(cfg.OutputFormats(), "|")+")")
	_ = cv.BindPFlag("output.format", cmd.PersistentFlags().Lookup("output-format"))

	cmd.PersistentFlags().StringP("output-query", "Q", cv.GetString("output.query"), "Filters output using JMESPath query (only JSON and YML formats)")
	_ = cv.BindPFlag("output.query", cmd.PersistentFlags().Lookup("output-query"))

	cmd.PersistentFlags().String("output-log-file", cv.GetString("output.log.file"), "Controls output file path")
	_ = cv.BindPFlag("output.log.file", cmd.PersistentFlags().Lookup("output-log-file"))

	cmd.PersistentFlags().String("output-log-mode", cv.GetString("output.log.mode"), "Controls where outputs and logs should be written to when format is \"text\" ("+(strings.Join(cfg.OutputLogModes(), "|")+")"))
	_ = cv.BindPFlag("output.log.mode", cmd.PersistentFlags().Lookup("output-log-mode"))

	cmd.PersistentFlags().StringSliceP("instance-url", "U", cv.GetStringSlice("instance.adhoc_url"), "Use only AEM instance(s) at ad-hoc specified URL(s)")
	_ = cv.BindPFlag("instance.adhoc_url", cmd.PersistentFlags().Lookup("instance-url"))

	cmd.PersistentFlags().StringSliceP("instance-id", "I", cv.GetStringSlice("instance.filter.id"), "Use only AEM instance(s) configured with the exact ID")
	_ = cv.BindPFlag("instance.filter.id", cmd.PersistentFlags().Lookup("instance-id"))

	cmd.PersistentFlags().BoolP("instance-author", "A", cv.GetBool("instance.filter.authors"), "Use only AEM author instance(s)")
	_ = cv.BindPFlag("instance.filter.authors", cmd.PersistentFlags().Lookup("instance-author"))

	cmd.PersistentFlags().BoolP("instance-publish", "P", cv.GetBool("instance.filter.publishes"), "Use only AEM publish instance(s)")
	_ = cv.BindPFlag("instance.filter.publishes", cmd.PersistentFlags().Lookup("instance-publish"))

	cmd.PersistentFlags().String("instance-processing", cv.GetString("instance.processing_mode"), "Controls processing mode for instances ("+(strings.Join(instance.ProcessingModes(), "|")+")"))
	_ = cv.BindPFlag("instance.processing_mode", cmd.PersistentFlags().Lookup("instance-processing"))
}
