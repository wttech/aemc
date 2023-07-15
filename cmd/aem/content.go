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
	cmd.AddCommand(c.contentMoveCmd())
	return cmd
}

func (c *CLI) contentCleanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "clean",
		Aliases: []string{"cln"},
		Short:   "Clean downloaded content",
		Run: func(cmd *cobra.Command, args []string) {
			rootPath, err := c.determineRootPath(cmd)
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
			var rootPath string
			if err == nil {
				rootPath, err = c.determineRootPath(cmd)
			}
			var filterPath string
			if err == nil {
				filterPath, err = c.determineFilterPath(cmd)
			}
			clean, _ := cmd.Flags().GetBool("clean")
			if err == nil {
				err = pkg.NewDownloader(c.aem.ContentOpts()).Download(instance.PackageManager(), rootPath, filterPath, clean)
			}
			if err != nil {
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
	cmd.Flags().Bool("clean", true, "Clean content after download")
	return cmd
}

func (c *CLI) contentMoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "move",
		Aliases: []string{"mv"},
		Short:   "Move content from one instance to another",
		Run: func(cmd *cobra.Command, args []string) {
			scrInstance, err := c.determineInstance(cmd, "src-instance-url", "src-instance-id", "unable to determine source instance")
			var descInstance *pkg.Instance
			if err == nil {
				descInstance, err = c.determineInstance(cmd, "desc-instance-url", "desc-instance-id", "unable to determine destination instance")
			}
			var filterPath string
			if err == nil {
				filterPath, err = c.determineFilterPath(cmd)
			}
			clean, _ := cmd.Flags().GetBool("clean")
			if err == nil {
				err = pkg.NewMover(c.aem.ContentOpts()).Move(scrInstance.PackageManager(), descInstance.PackageManager(), filterPath, clean)
			}
			if err != nil {
				c.Error(fmt.Errorf("content move failed: %w", err))
				return
			}
			c.Ok("content moved")
		},
	}
	cmd.Flags().String("src-instance-url", "", "Source instance URL")
	cmd.Flags().String("src-instance-id", "", "Source instance ID")
	cmd.MarkFlagsMutuallyExclusive("src-instance-url", "src-instance-id")
	cmd.Flags().String("desc-instance-url", "", "Destination instance URL")
	cmd.Flags().String("desc-instance-id", "", "Destination instance ID")
	cmd.MarkFlagsMutuallyExclusive("desc-instance-url", "desc-instance-id")
	cmd.Flags().String("filter-path", "", "Filter path")
	_ = cmd.MarkFlagRequired("filter-path")
	cmd.Flags().Bool("clean", true, "Clean content before move")
	return cmd
}

func (c *CLI) determineInstance(cmd *cobra.Command, urlParamName string, idParamName string, errorMsg string) (*pkg.Instance, error) {
	var instance *pkg.Instance
	if cmd.Flags().Lookup(urlParamName) != nil {
		url, err := cmd.Flags().GetString(urlParamName)
		if err == nil {
			instance, err = c.aem.InstanceManager().NewByURL(url)
		}
		return instance, nil
	}
	if cmd.Flags().Lookup(idParamName) != nil {
		id, err := cmd.Flags().GetString(idParamName)
		if err == nil {
			instance = c.aem.InstanceManager().NewByID(id)
		}
	}
	if instance == nil {
		return nil, fmt.Errorf(errorMsg)
	}
	return instance, nil
}

func (c *CLI) determineRootPath(cmd *cobra.Command) (string, error) {
	rootPath, err := cmd.Flags().GetString("root-path")
	if err == nil && !strings.Contains(rootPath, content.JcrRoot) {
		err = fmt.Errorf("root path '%s' does not contain '%s'", rootPath, content.JcrRoot)
	}
	return rootPath, err
}

func (c *CLI) determineFilterPath(cmd *cobra.Command) (string, error) {
	filterPath, err := cmd.Flags().GetString("filter-path")
	if err == nil && !strings.HasSuffix(filterPath, pkg.FilterXml) {
		err = fmt.Errorf("filter path '%s' does not end '%s'", filterPath, pkg.FilterXml)
	}
	return filterPath, err
}
