package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg/project"
	"strings"
)

func (c *CLI) projectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project",
		Short:   "Manage project files",
		Aliases: []string{"prj"},
	}
	cmd.AddCommand(c.projectInitCmd())
	cmd.AddCommand(c.projectScaffoldCmd())

	return cmd
}

const projectKindFlag = "project-kind"

func (c *CLI) projectScaffoldCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "scaffold",
		Aliases: []string{"setup"},
		Short:   "Scaffold required files in the project",
		Run: func(cmd *cobra.Command, args []string) {
			kindName, _ := cmd.Flags().GetString(projectKindFlag)
			kind, err := c.aem.Project().KindDetermine(kindName)
			if err != nil {
				c.Error(err)
				return
			}
			if kind == project.KindUnknown {
				c.Fail(fmt.Sprintf("project kind cannot be determined; specify it with flag '--%s=[%s]'", projectKindFlag, strings.Join(project.KindStrings(), "|")))
				return
			}

			changed, err := c.aem.Project().ScaffoldWithChanged(kind)
			if err != nil {
				c.Error(err)
				return
			}

			c.SetOutput("gettingStarted", c.aem.Project().ScaffoldGettingStarted())

			if changed {
				c.Changed("project files scaffolded")
			} else {
				c.Ok("project files already scaffolded")
			}
		},
	}
	cmd.Flags().String(projectKindFlag, project.KindAuto, fmt.Sprintf("Type of AEM to work with (%s)", strings.Join(project.KindStrings(), "|")))
	return cmd
}

func (c *CLI) projectInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "init",
		Aliases: []string{"initialize"},
		Short:   "Initializes AEMC in the project",
		Run: func(cmd *cobra.Command, args []string) {
			if !c.aem.Project().IsScaffolded() {
				c.Fail(fmt.Sprintf("project need to be scaffolded before running initialization"))
				return
			}

			changed := false

			c.SetOutput("gettingStarted", c.aem.Project().InitGettingStartedError())

			baseChanged, err := c.aem.BaseOpts().PrepareWithChanged()
			changed = changed || baseChanged
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("baseChanged", baseChanged)

			// Download and prepare vendor tools (including JDK and AEM SDK)
			vendorPrepared, err := c.aem.VendorManager().PrepareWithChanged()
			changed = changed || vendorPrepared
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("vendorPrepared", vendorPrepared)

			// Validate AEM instance files and prepared SDK
			if err := c.aem.InstanceManager().LocalOpts.Validate(); err != nil {
				c.Error(err)
				return
			}

			c.SetOutput("gettingStarted", c.aem.Project().InitGettingStartedSuccess())

			if changed {
				c.Changed("project initialized")
			} else {
				c.Ok("project already initialized")
			}
		},
	}
	return cmd
}
