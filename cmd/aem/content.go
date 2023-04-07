package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg/content"
)

func (c *CLI) contentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "content",
		Aliases: []string{"cnt"},
		Short:   "Manages content",
	}
	cmd.AddCommand(c.contentDownloadCmd())
	cmd.AddCommand(c.contentCleanCmd())
	cmd.AddCommand(c.contentMoveCmd())
	return cmd
}

func (c *CLI) contentCleanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "clean",
		Aliases: []string{"cln"},
		Short:   "Clean downloaded content",
		Run: func(cmd *cobra.Command, args []string) {
			rootPath, err := cmd.Flags().GetString("root-path")
			if err == nil {
				err = content.NewCleaner(c.aem.ContentOpts()).Clean(rootPath)
			}
			if err != nil {
				c.Error(fmt.Errorf("content clean failed: %w", err))
				return
			}
			c.Ok("content cleaned")
		},
	}
	cmd.Flags().String("root-path", "", "Root path")
	return cmd
}

func (c *CLI) contentMoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "move",
		Aliases: []string{"mv"},
		Short:   "Move content from one instance to another",
		Run: func(cmd *cobra.Command, args []string) {
			c.Ok("content moved")
		},
	}
	return cmd
}

func (c *CLI) contentDownloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "download",
		Aliases: []string{"dl"},
		Short:   "Download content from running instance",
		Run: func(cmd *cobra.Command, args []string) {
			c.Ok("content downloaded")
		},
	}
	return cmd
}
