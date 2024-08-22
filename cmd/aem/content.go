package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/content"
	"os"
	"path/filepath"
	"strings"
)

func (c *CLI) contentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "content",
		Aliases: []string{"cnt"},
		Short:   "Manages JCR content",
	}
	cmd.AddCommand(c.contentCleanCmd())
	cmd.AddCommand(c.contentPullCmd())
	cmd.AddCommand(c.contentPushCmd())
	cmd.AddCommand(c.contentDownloadCmd())
	cmd.AddCommand(c.contentCopyCmd())
	return cmd
}

func (c *CLI) contentCleanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "clean",
		Aliases: []string{"cln"},
		Short:   "Clean content",
		Run: func(cmd *cobra.Command, args []string) {
			dir, err := determineContentDir(cmd)
			if err != nil {
				c.Error(err)
				return
			}
			file, err := determineContentFile(cmd)
			if err != nil {
				c.Error(err)
				return
			}
			if dir != "" {
				if err = c.aem.ContentManager().CleanDir(dir); err != nil {
					c.Error(err)
					return
				}
				c.SetOutput("dir", dir)
			} else if file != "" {
				if err = c.aem.ContentManager().CleanFile(file); err != nil {
					c.Error(err)
					return
				}
				c.SetOutput("file", file)
			}
			c.Changed("content cleaned")
		},
	}
	cmd.Flags().StringP("dir", "d", "", "JCR root path")
	cmd.Flags().StringP("file", "f", "", "Local file path")
	cmd.Flags().StringP("path", "p", "", "JCR root path or local file path")
	cmd.MarkFlagsOneRequired("dir", "file", "path")
	return cmd
}

func (c *CLI) contentDownloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "download",
		Aliases: []string{"dl"},
		Short:   "Download content from running instance to local file",
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}
			pid, _ := cmd.Flags().GetString("pid")
			targetFile, _ := cmd.Flags().GetString("target-file")
			filterRoots := determineFilterRoots(cmd)
			filterFile, _ := cmd.Flags().GetString("filter-file")
			if err = instance.ContentManager().Download(targetFile, pkg.PackageCreateOpts{
				PID:         pid,
				FilterRoots: filterRoots,
				FilterFile:  filterFile,
			}); err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("file", targetFile)
			c.Changed("content downloaded")
		},
	}
	cmd.Flags().String("pid", "", "ID (group:name:version)'")
	cmd.Flags().StringP("target-file", "t", "", "Target file path for downloaded package")
	_ = cmd.MarkFlagRequired("target-file")
	cmd.Flags().StringSliceP("filter-roots", "r", []string{}, "Vault filter root paths")
	cmd.Flags().StringP("filter-file", "f", "", "Vault filter file path")
	cmd.MarkFlagsOneRequired("filter-roots", "filter-file")
	return cmd
}

func (c *CLI) contentPullCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pull",
		Aliases: []string{"pl", "sync"},
		Short:   "Pull content from running instance then unpack under JCR root directory",
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}
			clean, _ := cmd.Flags().GetBool("clean")
			replace, _ := cmd.Flags().GetBool("replace")
			dir, err := determineContentDir(cmd)
			if err != nil {
				c.Error(err)
				return
			}
			file, err := determineContentFile(cmd)
			if err != nil {
				c.Error(err)
				return
			}
			filterRoots := determineFilterRoots(cmd)
			filterFile, _ := cmd.Flags().GetString("filter-file")
			if dir != "" {
				if err = instance.ContentManager().PullDir(dir, clean, replace, pkg.PackageCreateOpts{
					FilterRoots: filterRoots,
					FilterFile:  filterFile,
				}); err != nil {
					c.Error(err)
					return
				}
				c.SetOutput("dir", dir)
			} else if file != "" {
				if err = instance.ContentManager().PullFile(file, clean, pkg.PackageCreateOpts{
					FilterRoots: filterRoots,
				}); err != nil {
					c.Error(err)
					return
				}
				c.SetOutput("file", file)
			}
			c.Changed("content synchronized")
		},
	}
	cmd.Flags().StringP("dir", "d", "", "JCR root path")
	cmd.Flags().String("file", "", "Local file path")
	cmd.Flags().StringP("path", "p", "", "JCR root path or local file path")
	cmd.MarkFlagsMutuallyExclusive("dir", "file", "path")
	cmd.Flags().StringSliceP("filter-roots", "r", []string{}, "Vault filter root paths")
	cmd.Flags().StringP("filter-file", "f", "", "Vault filter file path")
	cmd.MarkFlagsMutuallyExclusive("filter-roots", "filter-file")
	cmd.Flags().Bool("clean", false, "Normalize content after downloading")
	cmd.Flags().Bool("replace", false, "Replace content after downloading")
	return cmd
}

func (c *CLI) contentPushCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "push",
		Aliases: []string{"ps"},
		Short:   "Push content from JCR root directory or local file to running instance",
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}
			dir, err := determineContentDir(cmd)
			if err != nil {
				c.Error(err)
				return
			}
			file, err := determineContentFile(cmd)
			if err != nil {
				c.Error(err)
				return
			}
			path := dir
			if path == "" {
				path = file
			}
			clean, _ := cmd.Flags().GetBool("clean")
			filterRoots := determineFilterRoots(cmd)
			filterFileContent := determineFilterFileContent(cmd)
			if err = instance.ContentManager().Push(path, clean, pkg.PackageCreateOpts{
				FilterRoots:       filterRoots,
				FilterFileContent: filterFileContent,
			}); err != nil {
				c.Error(err)
				return
			}
			if dir != "" {
				c.SetOutput("dir", dir)
			} else if file != "" {
				c.SetOutput("file", file)
			}
			c.Changed("content pushed")
		},
	}
	cmd.Flags().StringP("dir", "d", "", "JCR root path")
	cmd.Flags().StringP("file", "f", "", "Local file path")
	cmd.Flags().StringP("path", "p", "", "JCR root path or local file path")
	cmd.MarkFlagsOneRequired("dir", "file", "path")
	cmd.Flags().Bool("clean", false, "Normalize content while uploading")
	return cmd
}

func (c *CLI) contentCopyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "copy",
		Aliases: []string{"cp"},
		Short:   "Copy content to another instance",
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}
			targetInstance, err := determineContentTargetInstance(cmd, c.aem.InstanceManager())
			if err != nil {
				c.Error(err)
				return
			}
			filterRoots := determineFilterRoots(cmd)
			filterFile, _ := cmd.Flags().GetString("filter-file")
			clean, _ := cmd.Flags().GetBool("clean")
			if err = instance.ContentManager().Copy(targetInstance, clean, pkg.PackageCreateOpts{
				FilterRoots: filterRoots,
				FilterFile:  filterFile,
			}); err != nil {
				c.Error(err)
				return
			}
			c.Changed("content copied")
		},
	}
	cmd.Flags().StringP("instance-target-url", "u", "", "Destination instance URL")
	cmd.Flags().StringP("instance-target-id", "i", "", "Destination instance ID")
	cmd.MarkFlagsOneRequired("instance-target-url", "instance-target-id")
	cmd.Flags().StringSliceP("filter-roots", "r", []string{}, "Vault filter root paths")
	cmd.Flags().StringP("filter-file", "f", "", "Vault filter file path")
	cmd.MarkFlagsOneRequired("filter-roots", "filter-file")
	cmd.Flags().Bool("clean", false, "Normalize content while copying")
	return cmd
}

func determineContentTargetInstance(cmd *cobra.Command, instanceManager *pkg.InstanceManager) (*pkg.Instance, error) {
	var instance *pkg.Instance
	url, _ := cmd.Flags().GetString("instance-target-url")
	if url != "" {
		instance, _ = instanceManager.NewByIDAndURL("remote_adhoc_target", url)
	}
	id, _ := cmd.Flags().GetString("instance-target-id")
	if id != "" {
		instance = instanceManager.NewByID(id)
	}
	if instance == nil {
		return nil, fmt.Errorf("missing 'instance-target-url' or 'instance-target-id'")
	}
	return instance, nil
}

func determineContentDir(cmd *cobra.Command) (string, error) {
	dir, _ := cmd.Flags().GetString("dir")
	if dir != "" && !strings.Contains(dir, content.JCRRoot) {
		return "", fmt.Errorf("content dir '%s' does not contain '%s'", dir, content.JCRRoot)
	}
	path, _ := cmd.Flags().GetString("path")
	if path != "" && !strings.Contains(path, content.JCRRoot) {
		return "", fmt.Errorf("content path '%s' does not contain '%s'", path, content.JCRRoot)
	}
	if path != "" && !pathx.Exists(path) {
		return "", fmt.Errorf("content path does not exist: %s", path)
	}
	if path != "" && pathx.IsDir(path) {
		return path, nil
	}
	return dir, nil
}

func determineContentFile(cmd *cobra.Command) (string, error) {
	file, _ := cmd.Flags().GetString("file")
	if file != "" && !strings.Contains(file, content.JCRRoot) {
		return "", fmt.Errorf("content file '%s' does not contain '%s'", file, content.JCRRoot)
	}
	path, _ := cmd.Flags().GetString("path")
	if path != "" && !strings.Contains(path, content.JCRRoot) {
		return "", fmt.Errorf("content path '%s' does not contain '%s'", path, content.JCRRoot)
	}
	if path != "" && !pathx.Exists(path) {
		return "", fmt.Errorf("content path does not exist: %s", path)
	}
	if path != "" && pathx.IsFile(path) {
		return path, nil
	}
	return file, nil
}

func determineFilterRoots(cmd *cobra.Command) []string {
	filterRoots, _ := cmd.Flags().GetStringSlice("filter-roots")
	if len(filterRoots) > 0 {
		return filterRoots
	}
	filterFile, _ := cmd.Flags().GetString("filter-file")
	if filterFile != "" {
		return nil
	}
	dir, _ := determineContentDir(cmd)
	if dir != "" {
		return []string{pkg.DetermineFilterRoot(dir)}
	}
	file, _ := determineContentFile(cmd)
	if file != "" {
		return []string{pkg.DetermineFilterRoot(file)}
	}
	return nil
}

func determineFilterFileContent(cmd *cobra.Command) string {
	file, _ := determineContentFile(cmd)
	if file == "" || !strings.HasSuffix(file, content.JCRContentFile) || content.IsContentFile(file) {
		return ""
	}

	dir := filepath.Dir(file)
	filterRoot := pkg.DetermineFilterRoot(file)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	filterFileContent := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"
	filterFileContent += "<workspaceFilter version=\"1.0\">\n"
	filterFileContent += fmt.Sprintf("<filter root=\"%s\">\n", filterRoot)
	for _, entry := range entries {
		if entry.Name() != content.JCRContentFile {
			jcrPath := pkg.DetermineFilterRoot(filepath.Join(dir, entry.Name()))
			filterFileContent += fmt.Sprintf("    <exclude pattern=\"%s(/.*)?\"/>\n", jcrPath)
		}
	}
	filterFileContent += "  </filter>\n"
	filterFileContent += "</workspaceFilter>\n"
	return filterFileContent
}
