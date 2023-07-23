package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg"
	"github.com/wttech/aemc/pkg/content"
	"strings"
)

func (c *CLI) contentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "content",
		Aliases: []string{"cnt"},
		Short:   "Manages content",
	}
	cmd.AddCommand(c.contentCleanCmd())
	cmd.AddCommand(c.contentDownloadCmd())
	cmd.AddCommand(c.contentCopyCmd())
	return cmd
}

func (c *CLI) contentCleanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "clean",
		Aliases: []string{"cln"},
		Short:   "Clean downloaded content",
		Run: func(cmd *cobra.Command, args []string) {
			rootPath, err := determineRootPath(cmd)
			if err != nil {
				c.Error(fmt.Errorf("content clean failed: %w", err))
				return
			}
			if err = content.NewCleaner(c.aem.ContentOpts()).Clean(rootPath); err != nil {
				c.Error(fmt.Errorf("content clean failed: %w", err))
				return
			}
			c.Ok("content cleaned")
		},
	}
	cmd.Flags().String("root-path", "", "Root path")
	_ = cmd.MarkFlagRequired("root-path")
	return cmd
}

func (c *CLI) contentDownloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "download",
		Aliases: []string{"dl"},
		Short:   "Download content from running instance",
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(fmt.Errorf("content download failed: %w", err))
				return
			}
			rootPath, err := determineRootPath(cmd)
			if err != nil {
				c.Error(fmt.Errorf("content download failed: %w", err))
				return
			}
			filterPath, err := determineFilterPath(cmd)
			if err != nil {
				c.Error(fmt.Errorf("content download failed: %w", err))
				return
			}
			onlyDownload, _ := cmd.Flags().GetBool("only-download")
			if err = pkg.NewDownloader(c.aem.ContentOpts()).DownloadContent(instance.PackageManager(), rootPath, filterPath, !onlyDownload); err != nil {
				c.Error(fmt.Errorf("content download failed: %w", err))
				return
			}
			c.Ok("content downloaded")
		},
	}
	cmd.Flags().String("root-path", "", "Root path")
	_ = cmd.MarkFlagRequired("root-path")
	cmd.Flags().String("filter-path", "", "Filter path")
	_ = cmd.MarkFlagRequired("filter-path")
	cmd.Flags().Bool("only-download", false, "Only download content")
	return cmd
}

func (c *CLI) contentCopyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "copy",
		Aliases: []string{"cp"},
		Short:   "Copy content from one instance to another",
		Run: func(cmd *cobra.Command, args []string) {
			scrInstance, err := determineInstance(cmd, c.aem.InstanceManager(), "src-instance-url", "src-instance-id", "unable to determine source instance")
			if err != nil {
				c.Error(fmt.Errorf("content copy failed: %w", err))
				return
			}
			destInstance, err := determineInstance(cmd, c.aem.InstanceManager(), "dest-instance-url", "dest-instance-id", "unable to determine destination instance")
			if err != nil {
				c.Error(fmt.Errorf("content copy failed: %w", err))
				return
			}
			filterPath, err := determineFilterPath(cmd)
			if err != nil {
				c.Error(fmt.Errorf("content copy failed: %w", err))
				return
			}
			onlyCopy, _ := cmd.Flags().GetBool("only-copy")
			if err = pkg.NewCopier(c.aem.ContentOpts()).Copy(scrInstance.PackageManager(), destInstance.PackageManager(), filterPath, !onlyCopy); err != nil {
				c.Error(fmt.Errorf("content copy failed: %w", err))
				return
			}
			c.Ok("content copied")
		},
	}
	cmd.Flags().String("src-instance-url", "", "Source instance URL")
	cmd.Flags().String("src-instance-id", "", "Source instance ID")
	cmd.MarkFlagsMutuallyExclusive("src-instance-url", "src-instance-id")
	cmd.Flags().String("dest-instance-url", "", "Destination instance URL")
	cmd.Flags().String("dest-instance-id", "", "Destination instance ID")
	cmd.MarkFlagsMutuallyExclusive("dest-instance-url", "dest-instance-id")
	cmd.Flags().String("filter-path", "", "Filter path")
	_ = cmd.MarkFlagRequired("filter-path")
	cmd.Flags().Bool("only-copy", false, "Only copy content")
	return cmd
}

func determineInstance(cmd *cobra.Command, instanceManager *pkg.InstanceManager, urlParamName string, idParamName string, errorMsg string) (*pkg.Instance, error) {
	var instance *pkg.Instance
	url, _ := cmd.Flags().GetString(urlParamName)
	if url != "" {
		instance, _ = instanceManager.NewByURL(url)
	}
	id, _ := cmd.Flags().GetString(idParamName)
	if id != "" {
		instance = instanceManager.NewByID(id)
	}
	if instance == nil {
		return nil, fmt.Errorf(errorMsg)
	}
	return instance, nil
}

func determineRootPath(cmd *cobra.Command) (string, error) {
	rootPath, _ := cmd.Flags().GetString("root-path")
	if !strings.Contains(rootPath, content.JcrRoot) {
		return "", fmt.Errorf("root path '%s' does not contain '%s'", rootPath, content.JcrRoot)
	}
	return rootPath, nil
}

func determineFilterPath(cmd *cobra.Command) (string, error) {
	filterPath, _ := cmd.Flags().GetString("filter-path")
	if !strings.HasSuffix(filterPath, pkg.FilterXml) {
		return "", fmt.Errorf("filter path '%s' does not end '%s'", filterPath, pkg.FilterXml)
	}
	return filterPath, nil
}
