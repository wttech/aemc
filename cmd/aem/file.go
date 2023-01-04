package main

import (
	"github.com/spf13/cobra"
	f "github.com/wttech/aemc/pkg/file"
	"strings"
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
			opts := f.DownloadOpts{}

			url, _ := cmd.Flags().GetString("url")
			opts.Url = url

			file, _ := cmd.Flags().GetString("file")
			opts.File = file

			authBasic, _ := cmd.Flags().GetString("auth-basic")
			if len(authBasic) > 0 {
				parts := strings.Split(authBasic, ":")
				opts.AuthBasicUser = parts[0]
				opts.AuthBasicPassword = parts[1]
			}
			authToken, _ := cmd.Flags().GetString("auth-token")
			opts.AuthToken = authToken

			c.SetOutput("url", url)
			c.SetOutput("file", file)

			changed, err := f.DownloadWithOpts(opts)
			if err != nil {
				c.Error(err)
				return
			}
			if changed {
				c.Changed("file downloaded")
			} else {
				c.Ok("file already downloaded (up-to-date)")
			}
		},
	}
	cmd.Flags().String("url", "", "Source file URL")
	_ = cmd.MarkFlagRequired("url")
	cmd.Flags().String("file", "", "Destination file path")
	_ = cmd.MarkFlagRequired("file")
	cmd.Flags().String("auth-basic", "", "Basic Authorization (in format 'user:password')")
	cmd.Flags().String("auth-token", "", "Token Authorization")
	return cmd
}
