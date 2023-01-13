package main

import (
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg/common/pathx"
)

func (c *CLI) appCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "application",
		Aliases: []string{"app"},
		Short:   "Application build utilities",
	}
	cmd.AddCommand(c.appBuildCmd())
	return cmd
}

func (c *CLI) appBuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build application only when needed",
		Run: func(cmd *cobra.Command, args []string) {
			command, _ := cmd.Flags().GetString("command")
			file, _ := cmd.Flags().GetString("file")
			fileGlobbed, err := pathx.GlobOne(file)
			if err != nil {
				c.Error(err)
				return
			}
			appManager := c.aem.AppManager()

			sourcePaths, _ := cmd.Flags().GetStringSlice("sources")
			sourcesIgnoredExtra, _ := cmd.Flags().GetStringSlice("sources-ignored")
			appManager.SourcesIgnored = lo.Uniq(append(appManager.SourcesIgnored, sourcesIgnoredExtra...))

			changed, err := appManager.BuildWithChanged(command, fileGlobbed, sourcePaths)
			if err != nil {
				c.Error(err)
				return
			}

			c.SetOutput("command", command)
			c.SetOutput("file", fileGlobbed)
			c.SetOutput("sources", sourcePaths)

			if changed {
				c.Changed("build executed")
			} else {
				c.Ok("build not executed (up-to-date)")
			}
		},
	}
	cmd.Flags().String("command", "", "AEM application build command")
	_ = cmd.MarkFlagRequired("command")
	cmd.Flags().String("file", "", "Path or pattern for built file")
	_ = cmd.MarkFlagRequired("file")
	cmd.Flags().StringSlice("sources", []string{}, "Source directories")
	_ = cmd.MarkFlagRequired("sources")
	cmd.Flags().StringSlice("sources-ignored", []string{}, "Ignored sources patterns (git-ignore style)")
	return cmd
}
