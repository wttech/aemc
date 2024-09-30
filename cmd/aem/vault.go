package main

import (
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg"
	"os"
)

func (c *CLI) vaultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vlt",
		Short: "Executes Vault-Cli commands",
		Run: func(cmd *cobra.Command, args []string) {
			vaultCli := pkg.NewVaultCli(c.aem)
			vaultCliArgs := os.Args[1:]
			_ = vaultCli.CommandShell(vaultCliArgs)
		},
		Args: cobra.ArbitraryArgs,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
	}
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		aem := pkg.NewAEM(c.config)
		vaultCli := pkg.NewVaultCli(aem)
		_ = vaultCli.CommandShell(args)
	})
	return cmd
}
