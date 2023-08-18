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
		Use:   "status",
		Short: "Get status of Global Trust Store",
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}

			result, err := instance.GlobalTrustStore().Status()

			if err != nil {
				c.Error(err)
				return
			}

			if result.Created == nil {
				c.SetOutput("Global Trust Store", result)
			}
		},
	}

	return cmd
}

func (c *CLI) gtsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{},
		Short:   "Create Global Trust Store",
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}

			trustStorePassword, _ := cmd.Flags().GetString("password")

			changed, err := instance.GlobalTrustStore().Create(trustStorePassword)

			if err != nil {
				c.Error(err)
				return
			}

			if changed {
				c.Changed("Global trust store created")
			} else {
				c.Ok("Global trust store already exists")
			}

			cmd.Flags().String("password", "", "Password to Global Trust Store")
			_ = cmd.MarkFlagRequired("password")
		},
	}

	return cmd
}

func (c *CLI) gtsCertCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "certificate",
		Short: "Manage Global Trust Store Certificates",
	}
	cmd.AddCommand(c.gtsCertAddCmd())
	cmd.AddCommand(c.gtsCertRemoveCmd())
	cmd.AddCommand(c.gtsCertReadCmd())
	return cmd

}

func (c *CLI) gtsCertAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add",
		Aliases: []string{},
		Short:   "Add Certificate to Global Trust Store",
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}

			certificateFilePath, _ := cmd.Flags().GetString("path")

			certificate, changed, err := instance.GlobalTrustStore().AddCertificate(certificateFilePath)

			if err != nil {
				c.Error(err)
				return
			}

			c.SetOutput("Certificate", certificate.Alias)

			if changed {
				c.Changed("Certificate added")
			} else {
				c.Ok("Certificate already exists")
			}
		},
	}

	cmd.Flags().String("path", "", "Certificate file path (PEM|DER format)")
	_ = cmd.MarkFlagRequired("path")

	return cmd
}

func (c *CLI) gtsCertRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove",
		Aliases: []string{},
		Short:   "Remove Certificate from Global Trust Store",
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}

			certifiacteAlias, _ := cmd.Flags().GetString("alias")
			changed, err := instance.GlobalTrustStore().RemoveCertificate(certifiacteAlias)

			if err != nil {
				c.Error(err)
				return
			}

			if changed {
				c.Changed("Certificate removed")
			} else {
				c.Ok("Certificate not found")
			}
		},
	}

	cmd.Flags().String("alias", "", "Certificate Alias")
	_ = cmd.MarkFlagRequired("alias")

	return cmd
}

func (c *CLI) gtsCertReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "read",
		Aliases: []string{},
		Short:   "Read certificate from Global Trust Store",
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}

			certifiacteAlias, _ := cmd.Flags().GetString("alias")
			certificate, err := instance.GlobalTrustStore().ReadCertificate(certifiacteAlias)

			if err != nil {
				c.Error(err)
				return
			}

			c.SetOutput("Read", certificate)
		},
	}

	cmd.Flags().String("alias", "", "Certificate Alias")
	_ = cmd.MarkFlagRequired("alias")

	return cmd
}
