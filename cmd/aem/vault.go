package main

import (
	"github.com/spf13/cobra"
)

func (c *CLI) vaultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vlt",
		Short: "Executes Vault commands",
		Run: func(cmd *cobra.Command, args []string) {
			if err := c.aem.VendorManager().VaultCLI().CommandShell(args); err != nil {
				c.Error(err)
				return
			}
		},
		Args: cobra.ArbitraryArgs,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
	}
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		_ = c.aem.VendorManager().VaultCLI().CommandShell(args)
	})
	return cmd
}
