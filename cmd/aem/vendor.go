package main

import (
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
		Short:   "List all prepared vendors",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
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

			c.Ok("vendors listed")
		},
	}
	replActivationFlags(cmd)

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
	replActivationFlags(cmd)
	return cmd
}
