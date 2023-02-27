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
		Aliases: []string{"prj"},
		Short:   "Manages project",
	}
	cmd.AddCommand(c.projectInitCmd())
	return cmd
}

func (c *CLI) projectInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "initialize",
		Aliases: []string{"init"},
		Short:   "Initializes project files and configuration",
		Run: func(cmd *cobra.Command, args []string) {
			if err := c.project.EnsureDirs(); err != nil {
				c.Error(err)
				return
			}
			kindName, _ := cmd.Flags().GetString("kind")
			kind, err := c.project.KindDetermine(kindName)
			if err != nil {
				c.Error(err)
				return
			}
			if kind == project.KindUnknown {
				c.Fail(fmt.Sprintf("project kind cannot be determined; specify it with flag '--kind=[%s]'", strings.Join(project.KindStrings(), "|")))
				return
			}
			changed, err := c.project.InitializeWithChanged(kind)
			if err != nil {
				c.Error(err)
				return
			}
			gettingStarted, err := c.project.GettingStarted()
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("gettingStarted", gettingStarted)
			if changed {
				c.Changed("project initialized")
			} else {
				c.Ok("project already initialized")
			}
		},
	}
	cmd.Flags().String("kind", project.KindAuto, fmt.Sprintf("Type of AEM to work with (%s)", strings.Join(project.KindStrings(), "|")))
	return cmd
}
