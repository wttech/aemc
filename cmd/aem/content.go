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
		Short:   "Manages JCR content",
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
			contentRootPath, err := determineContentRootPath(cmd)
			if err != nil {
				c.Error(fmt.Errorf("content clean failed: %w", err))
				return
			}
			if err = content.NewCleaner(c.aem.ContentOpts()).Clean(contentRootPath); err != nil {
				c.Error(fmt.Errorf("content clean failed: %w", err))
				return
			}
			c.Ok("content cleaned")
		},
	}
	cmd.Flags().StringP("content-root-path", "R", "", "Content root path on file system")
	_ = cmd.MarkFlagRequired("content-root-path")
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
			pid, _ := cmd.Flags().GetString("pid")
			contentRootPath, err := determineContentRootPath(cmd)
			if err != nil {
				c.Error(fmt.Errorf("content download failed: %w", err))
				return
			}
			rootPaths, _ := cmd.Flags().GetStringSlice("root-path")
			filterFile, err := determineFilterFile(cmd)
			if err != nil {
				c.Error(fmt.Errorf("content download failed: %w", err))
				return
			}
			onlyDownload, _ := cmd.Flags().GetBool("only-download")
			onlyPackage, _ := cmd.Flags().GetBool("only-package")
			if err = pkg.NewDownloader(c.aem.ContentOpts()).DownloadContent(instance.PackageManager(), pid, contentRootPath, rootPaths, filterFile, !onlyDownload, !onlyPackage); err != nil {
				c.Error(fmt.Errorf("content download failed: %w", err))
				return
			}
			c.Ok("content downloaded")
		},
	}
	cmd.Flags().String("pid", "", "ID (group:name:version)'")
	cmd.Flags().StringP("content-root-path", "R", "", "Content root path on file system")
	_ = cmd.MarkFlagRequired("content-root-path")
	cmd.Flags().StringSliceP("root-path", "r", []string{}, "Filter root path(s) on AEM repository")
	cmd.Flags().StringP("filter-file", "f", "", "Local filter file on file system")
	cmd.MarkFlagsMutuallyExclusive("root-path", "filter-file")
	cmd.Flags().Bool("only-download", false, "Only download content")
	cmd.Flags().Bool("only-package", false, "Only download package")
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
			rootPaths, _ := cmd.Flags().GetStringSlice("root-path")
			filterFile, err := determineFilterFile(cmd)
			if err != nil {
				c.Error(fmt.Errorf("content copy failed: %w", err))
				return
			}
			onlyCopy, _ := cmd.Flags().GetBool("only-copy")
			if err = pkg.NewCopier(c.aem.ContentOpts()).Copy(scrInstance.PackageManager(), destInstance.PackageManager(), "", rootPaths, filterFile, !onlyCopy); err != nil {
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
	cmd.Flags().StringSliceP("root-path", "r", []string{}, "Filter root path(s) on AEM repository")
	cmd.Flags().StringP("filter-file", "f", "", "Local filter file on file system")
	cmd.MarkFlagsMutuallyExclusive("root-path", "filter-file")
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

func determineContentRootPath(cmd *cobra.Command) (string, error) {
	rootPath, _ := cmd.Flags().GetString("content-root-path")
	onlyPackage, _ := cmd.Flags().GetBool("only-package")
	if !onlyPackage && !strings.Contains(rootPath, content.JcrRoot) {
		return "", fmt.Errorf("root path '%s' does not contain '%s'", rootPath, content.JcrRoot)
	}
	return rootPath, nil
}

func determineFilterFile(cmd *cobra.Command) (string, error) {
	filterPath, _ := cmd.Flags().GetString("filter-file")
	if filterPath != "" && !strings.HasSuffix(filterPath, pkg.FilterXml) {
		return "", fmt.Errorf("filter path '%s' does not end '%s'", filterPath, pkg.FilterXml)
	}
	return filterPath, nil
}
