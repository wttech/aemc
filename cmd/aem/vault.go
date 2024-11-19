package main

import (
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg"
)

func (c *CLI) vaultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "vlt",
		Short:  "Executes Vault commands",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			if err := c.aem.VendorManager().VaultCLI().CommandShell(args); err != nil {
				c.Fail("command failed")
				return
			}
			c.Ok("command run")
		},
		Args: cobra.ArbitraryArgs,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
	}
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		aem := pkg.NewAEM(c.config) // c.onStart() not yet called
		_ = aem.VendorManager().VaultCLI().CommandShell(args[1:])
	})
	return cmd
}
