package main

import (
	"github.com/spf13/cobra"
	"os"
)

func (c *CLI) vaultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vlt",
		Short: "Executes Vault commands",
		Run: func(cmd *cobra.Command, args []string) {
			argsWithVlt := os.Args[1:]                     // TODO why not 'args' from the Run function?
			_ = c.aem.VaultCLI().CommandShell(argsWithVlt) // TODO proper error handling
		},
		Args: cobra.ArbitraryArgs,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
	}
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		_ = c.aem.VaultCLI().CommandShell(args)
	})
	return cmd
}
