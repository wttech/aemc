package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (c *CLI) userCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "user",
		Short:   "User management",
		Aliases: []string{"usr"},
	}
	cmd.AddCommand(c.userKeyStore())
	cmd.AddCommand(c.userKey())
	cmd.AddCommand(c.userPassword())
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

func (c *CLI) userKey() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "key",
		Short:   "Private keys management",
		Aliases: []string{"keys"},
	}
	cmd.AddCommand(c.userKeyAdd())
	cmd.AddCommand(c.userKeyDelete())
	return cmd
}

func (c *CLI) userPassword() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "password",
		Short:   "Password management",
		Aliases: []string{"pwd"},
	}
	cmd.AddCommand(c.UserPasswordSet())
	return cmd
}

func (c *CLI) KeystoreStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status",
		Short:   "Get status of a user's keystore",
		Aliases: []string{"show", "get", "read", "describe", "ls"},
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}

			id, _ := cmd.Flags().GetString("id")
			scope, _ := cmd.Flags().GetString("scope")

			result, err := instance.Auth().UserManager().Keystore().Status(scope, id)

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
	return cmd
}

func (c *CLI) KeystoreCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create user's keystore",
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
			changed, err := instance.Auth().UserManager().Keystore().Create(scope, id, password)

			if err != nil {
				c.Error(err)
				return
			}

			if changed {
				c.Changed("User keystore created")
			} else {
				c.Ok("User keystore already exists")
			}
		},
	}

	cmd.Flags().String("id", "", "user id")
	_ = cmd.MarkFlagRequired("id")
	cmd.Flags().String("scope", "", "user scope")
	cmd.Flags().String("keystore-password", "", "keystore password")
	_ = cmd.MarkFlagRequired("keystore-password")
	return cmd
}

func (c *CLI) userKeyAdd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add",
		Short:   "Add user's private key to their keystore",
		Aliases: []string{"create", "new"},
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}

			changed, err := instance.Auth().UserManager().Keystore().AddKey(
				cmd.Flag("scope").Value.String(),
				cmd.Flag("id").Value.String(),
				cmd.Flag("keystore-file").Value.String(),
				cmd.Flag("keystore-password").Value.String(),
				cmd.Flag("key-alias").Value.String(),
				cmd.Flag("key-password").Value.String(),
				cmd.Flag("new-alias").Value.String(),
			)

			if err != nil {
				c.Error(err)
				return
			}
			if changed {
				c.Changed("User key added")
			} else {
				c.Ok("User key already exists")
			}
		},
	}

	cmd.Flags().String("id", "", "user id")
	_ = cmd.MarkFlagRequired("id")
	cmd.Flags().String("scope", "", "user scope")
	cmd.Flags().String("keystore-file", "", "path to keystore file")
	_ = cmd.MarkFlagRequired("keystore-file")
	cmd.Flags().String("keystore-password", "", "keystore password")
	_ = cmd.MarkFlagRequired("keystore-password")
	cmd.Flags().String("key-alias", "", "key alias")
	_ = cmd.MarkFlagRequired("key-alias")
	cmd.Flags().String("key-password", "", "key password")
	cmd.Flags().String("new-alias", "", "new key alias (optional)")

	return cmd
}

func (c *CLI) userKeyDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete user's private key from their keystore",
		Aliases: []string{"remove", "rm"},
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}

			changed, err := instance.Auth().UserManager().Keystore().DeleteKey(
				cmd.Flag("scope").Value.String(),
				cmd.Flag("id").Value.String(),
				cmd.Flag("key-alias").Value.String(),
			)

			if err != nil {
				c.Error(err)
				return
			}
			if changed {
				c.Changed("User key deleted")
			} else {
				c.Ok("User key does not exist")
			}
		},
	}

	cmd.Flags().String("id", "", "user id")
	_ = cmd.MarkFlagRequired("id")
	cmd.Flags().String("scope", "", "user scope")
	cmd.Flags().String("key-alias", "", "key alias")
	_ = cmd.MarkFlagRequired("key-alias")

	return cmd
}

func (c *CLI) UserPasswordSet() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set",
		Short:   "Set user password. Password is read from input.",
		Aliases: []string{"update", "change"},
		Run: func(cmd *cobra.Command, args []string) {
			instances, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}

			id, _ := cmd.Flags().GetString("id")
			scope, _ := cmd.Flags().GetString("scope")

			var password string
			if err := c.ReadInput(&password); err != nil {
				c.Fail(fmt.Sprintf("error reading password from input: %s", err))
				return
			}

			changed, err := instances.Auth().UserManager().SetPassword(scope, id, password)
			if err != nil {
				c.Error(err)
				return
			}

			if changed {
				c.Changed("User password changed")
			} else {
				c.Ok("User password already set")
			}
		},
	}

	cmd.Flags().String("id", "", "user id")
	_ = cmd.MarkFlagRequired("id")
	cmd.Flags().String("scope", "", "user scope")

	return cmd
}
