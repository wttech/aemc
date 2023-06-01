package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg/project"
	"strings"
)

const projectKindFlag = "project-kind"

func (c *CLI) initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "init",
		Aliases: []string{"initialize"},
		Short:   "Initializes project files and configuration",
		Run: func(cmd *cobra.Command, args []string) {
			if err := c.aem.Project().EnsureDirs(); err != nil {
				c.Error(err)
				return
			}
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
			changed, err := c.aem.Project().InitializeWithChanged(kind)
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
				c.Changed("project initialized")
			} else {
				c.Ok("project already initialized")
			}
		},
	}
	cmd.Flags().String(projectKindFlag, project.KindAuto, fmt.Sprintf("Type of AEM to work with (%s)", strings.Join(project.KindStrings(), "|")))
	return cmd
}
