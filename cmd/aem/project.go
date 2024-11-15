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
		Short:   "Initializes AEMC in the project",
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
			gettingStarted, err := c.aem.Project().GettingStarted(kind)
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

func (c *CLI) prepareCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "prepare",
		Aliases: []string{"prep"},
		Short:   "Prepare vendor tools",
		Run: func(cmd *cobra.Command, args []string) {
			if err := c.aem.VendorManager().Prepare(); err != nil {
				c.Error(err)
				return
			}

			javaHome, err := c.aem.VendorManager().JavaManager().FindHomeDir()
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("javaHome", javaHome)

			javaExecutable, err := c.aem.VendorManager().JavaManager().Executable()
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("javaExecutable", javaExecutable)

			vaultJar := c.aem.VendorManager().VaultCLI().JarFile()
			c.setOutput("vaultJar", vaultJar)

			oakRunJar := c.aem.VendorManager().OakRun().JarFile()
			c.setOutput("oakRunJar", oakRunJar)

			c.SetOutput("prepared", true)
			c.Ok("prepared vendor tools")

			/* TODO implement if possible
			if changed {
				c.Changed("project prepared")
			} else {
				c.Ok("project already prepared")
			}
			*/
		},
	}
	return cmd
}
