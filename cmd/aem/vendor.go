package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (c *CLI) vendorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "vendor",
		Short:   "Supportive tools management",
		Aliases: []string{"ven"},
	}
	cmd.AddCommand(c.vendorListCmd())
	cmd.AddCommand(c.vendorPrepareCmd())

	return cmd
}

func (c *CLI) vendorListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List vendor tools available",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			errored := false

			javaHome, err := c.aem.VendorManager().JavaManager().FindHomeDir()
			if err != nil {
				errored = true
				log.Warnf("java home not available: %s", err)
			}
			c.SetOutput("javaHome", javaHome)

			javaExecutable, err := c.aem.VendorManager().JavaManager().Executable()
			if err != nil {
				errored = true
				log.Warnf("java executable not available: %s", err)
			}
			c.SetOutput("javaExecutable", javaExecutable)

			vaultJar := c.aem.VendorManager().VaultCLI().JarFile()
			c.setOutput("vaultJar", vaultJar)

			oakRunJar := c.aem.VendorManager().OakRun().JarFile()
			c.setOutput("oakRunJar", oakRunJar)

			if errored {
				c.Fail("vendor tool listed with errors")
			} else {
				c.Ok("vendor tools listed")
			}
		},
	}
	return cmd
}

func (c *CLI) vendorPrepareCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "prepare",
		Short:   "Prepare vendor tools",
		Aliases: []string{"prep", "download", "dw"},
		Run: func(cmd *cobra.Command, args []string) {
			changed, err := c.aem.VendorManager().PrepareWithChanged()
			if err != nil {
				c.Error(err)
				return
			}

			if changed {
				c.Changed("vendor tools prepared")
			} else {
				c.Ok("vendor tools already prepared")
			}
		},
	}
	return cmd
}