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

func (c *CLI) projectInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "init",
		Aliases: []string{"initialize"},
		Short:   "Initializes AEMC in the project",
		Run: func(cmd *cobra.Command, args []string) {
			if !c.aem.Project().IsScaffolded() {
				c.Fail(fmt.Sprintf("project need to be set up before running initialization"))
				return
			}

			vendorPrepared, err := c.aem.VendorManager().PrepareWithChanged()
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("vendorPrepared", vendorPrepared)

			gettingStarted, err := c.aem.Project().GettingStarted()
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("gettingStarted", gettingStarted)

			if vendorPrepared {
				c.Changed("initialized")
			} else {
				c.Ok("nothing to initialize")
			}
		},
	}
	cmd.Flags().String(projectKindFlag, project.KindAuto, fmt.Sprintf("Type of AEM to work with (%s)", strings.Join(project.KindStrings(), "|")))
	return cmd
}

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

			gettingStarted, err := c.aem.Project().GettingStarted()
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("gettingStarted", gettingStarted)

			if changed {
				c.Changed("project files scaffolded")
			} else {
				c.Ok("project files already scaffolded")
			}
		},
	}
	return cmd
}
