package main

import (
	"github.com/spf13/cobra"
)

func (c *CLI) gtsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gts",
		Short: "Communicate with Global Trust Store",
	}
	cmd.AddCommand(c.gtsStatusCmd())
	cmd.AddCommand(c.gtsCreateCmd())
	cmd.AddCommand(c.gtsCertCmd())
	return cmd
}

func (c *CLI) gtsStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status",
		Short:   "Get status of Global Trust Store",
		Aliases: []string{"show", "get", "read", "describe", "ls"},
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}

			result, err := instance.GTS().Status()

			if err != nil {
				c.Error(err)
				return
			}

			c.Ok("Global Trust Store status")
			c.SetOutput("status", result)
		},
	}

	return cmd
}

func (c *CLI) gtsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"make", "new"},
		Short:   "Create Global Trust Store",
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}

			trustStorePassword, _ := cmd.Flags().GetString("password")

			changed, err := instance.GTS().Create(trustStorePassword)

			if err != nil {
				c.Error(err)
				return
			}

			if changed {
				c.Changed("Global Trust Store created")
			} else {
				c.Ok("Global Trust Store already exists")
			}
		},
	}

	cmd.Flags().String("password", "", "password to Global Trust Store")
	_ = cmd.MarkFlagRequired("password")

	return cmd
}

func (c *CLI) gtsCertCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "certificate",
		Aliases: []string{"cert", "crt"},
		Short:   "Manage Global Trust Store certificates",
	}
	cmd.AddCommand(c.gtsCertAddCmd())
	cmd.AddCommand(c.gtsCertRemoveCmd())
	cmd.AddCommand(c.gtsCertReadCmd())
	return cmd

}

func (c *CLI) gtsCertAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add",
		Aliases: []string{"push", "install"},
		Short:   "Add cert to Global Trust Store",
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}

			certificateFilePath, _ := cmd.Flags().GetString("path")

			certificate, changed, err := instance.GTS().AddCertificate(certificateFilePath)

			if err != nil {
				c.Error(err)
				return
			}

			c.SetOutput("added", certificate.Alias)

			if changed {
				c.Changed("certificate added")
			} else {
				c.Ok("certificate already exists")
			}
		},
	}

	cmd.Flags().String("path", "", "file path (PEM|DER format)")
	_ = cmd.MarkFlagRequired("path")

	return cmd
}

func (c *CLI) gtsCertRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove",
		Aliases: []string{"del", "destroy", "rm"},
		Short:   "Remove certificate from Global Trust Store",
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}

			alias, _ := cmd.Flags().GetString("alias")
			changed, err := instance.GTS().RemoveCertificate(alias)

			if err != nil {
				c.Error(err)
				return
			}

			if changed {
				c.Changed("certificate removed")
				c.SetOutput("removed", alias)
			} else {
				c.Ok("certificate not found")
			}
		},
	}

	cmd.Flags().String("alias", "", "alias")
	_ = cmd.MarkFlagRequired("alias")

	return cmd
}

func (c *CLI) gtsCertReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "read",
		Aliases: []string{"cat", "show", "get", "describe", "find"},
		Short:   "Read certificate from Global Trust Store",
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}

			alias, _ := cmd.Flags().GetString("alias")
			certificate, err := instance.GTS().ReadCertificate(alias)

			if err != nil {
				c.Error(err)
				return
			}

			if certificate != nil {
				c.Ok("certificate found")
				c.SetOutput("certificate", certificate)
				return
			}

			c.Ok("certificate not found")
		},
	}

	cmd.Flags().String("alias", "", "alias")
	_ = cmd.MarkFlagRequired("alias")

	return cmd
}
