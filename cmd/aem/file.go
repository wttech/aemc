package main

import (
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg/common/filex"
	"github.com/wttech/aemc/pkg/common/httpx"
	"strings"
)

func (c *CLI) fileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file",
		Short: "File operation utilities",
	}
	cmd.AddCommand(c.fileDownloadCmd())
	cmd.AddCommand(c.fileArchiveCmd())
	cmd.AddCommand(c.fileUnarchiveCmd())
	return cmd
}

func (c *CLI) fileDownloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "download",
		Aliases: []string{"dwn", "get"},
		Short:   "Download file from URL",
		Run: func(cmd *cobra.Command, args []string) {
			opts := httpx.DownloadOpts{}

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

			changed, err := httpx.DownloadWithChanged(opts)
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

func (c *CLI) fileUnarchiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "unarchive",
		Aliases: []string{"unarch", "unzip", "extract", "decompress"},
		Short:   "Unarchive file",
		Run: func(cmd *cobra.Command, args []string) {
			sourceFile, _ := cmd.Flags().GetString("source-file")
			targetDir, _ := cmd.Flags().GetString("target-dir")

			c.SetOutput("sourceFile", sourceFile)
			c.SetOutput("targetDir", targetDir)

			changed, err := filex.UnarchiveWithChanged(sourceFile, targetDir)
			if err != nil {
				c.Error(err)
				return
			}
			if changed {
				c.Changed("file unarchived")
			} else {
				c.Ok("file already unarchived (up-to-date)")
			}
		},
	}
	cmd.Flags().String("source-file", "", "Source archive file path (zip/tar.gz/...)")
	_ = cmd.MarkFlagRequired("source-file")

	cmd.Flags().String("target-dir", "", "Target directory path for unarchived files")
	_ = cmd.MarkFlagRequired("target-dir")
	return cmd
}

func (c *CLI) fileArchiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "archive",
		Aliases: []string{"arch", "zip", "compact", "compress"},
		Short:   "Archive directory",
		Run: func(cmd *cobra.Command, args []string) {
			sourceDir, _ := cmd.Flags().GetString("source-dir")
			targetFile, _ := cmd.Flags().GetString("target-file")

			c.SetOutput("sourceDir", sourceDir)
			c.SetOutput("targetFile", targetFile)

			changed, err := filex.ArchiveWithChanged(sourceDir, targetFile)
			if err != nil {
				c.Error(err)
				return
			}
			if changed {
				c.Changed("directory archived")
			} else {
				c.Ok("directory already archived (up-to-date)")
			}
		},
	}
	cmd.Flags().String("source-dir", "", "Source directory with files to archive")
	_ = cmd.MarkFlagRequired("source-dir")

	cmd.Flags().String("target-file", "", "Target archive file path (zip/tar.gz/...)")
	_ = cmd.MarkFlagRequired("target-file")

	return cmd
}
