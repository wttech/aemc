package main

import "github.com/spf13/cobra"

func (c *CLI) userCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "user",
		Short:   "User management",
		Aliases: []string{"usr"},
	}
	cmd.AddCommand(c.userKeyStore())
	return cmd
}

func (c *CLI) userKeyStore() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "keystore",
		Short:   "Keystore management",
		Aliases: []string{"ks"},
	}
	cmd.AddCommand(c.KeystoreStatus())
	cmd.AddCommand(c.KeystoreCreate())
	return cmd
}

func (c *CLI) KeystoreStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status",
		Short:   "Get status of keystore",
		Aliases: []string{"show", "get", "read", "describe", "ls"},
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}

			id, _ := cmd.Flags().GetString("id")
			scope, _ := cmd.Flags().GetString("scope")

			result, err := instance.Auth().UserManager().KeystoreStatus(scope, id)

			if err != nil {
				c.Error(err)
				return
			}

			c.Ok("User Keystore status")
			c.SetOutput("status", result)
		},
	}
	cmd.Flags().String("id", "", "user id")
	_ = cmd.MarkFlagRequired("id")
	cmd.Flags().String("scope", "", "user scope")
	_ = cmd.MarkFlagRequired("scope")
	return cmd
}

func (c *CLI) KeystoreCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create user Keystore",
		Aliases: []string{"make", "new"},
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}

			id, _ := cmd.Flags().GetString("id")
			scope, _ := cmd.Flags().GetString("scope")
			password, _ := cmd.Flags().GetString("keystore-password")
			changed, err := instance.Auth().UserManager().KeystoreCreate(scope, id, password)

			if err != nil {
				c.Error(err)
				return
			}

			if changed {
				c.Changed("User Keystore created")
			} else {
				c.Ok("User Keystore already exists")
			}
		},
	}

	cmd.Flags().String("id", "", "user id")
	_ = cmd.MarkFlagRequired("id")
	cmd.Flags().String("scope", "", "user scope")
	_ = cmd.MarkFlagRequired("scope")
	cmd.Flags().String("keystore-password", "", "keystore password")
	_ = cmd.MarkFlagRequired("keystore-password")
	return cmd
}
