package main

import (
	"github.com/spf13/cobra"
)

func (c *CLI) fileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file",
		Short: "File operation utilities",
	}
	cmd.AddCommand(c.fileDownloadCmd())
	return cmd
}

func (c *CLI) fileDownloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "download",
		Aliases: []string{"dwn", "get"},
		Short:   "Download file from URL",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO impl
			c.Ok("file downloaded")
		},
	}
	return cmd
}
