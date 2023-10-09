package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg"
	"github.com/wttech/aemc/pkg/common/mapsx"
)

func (c *CLI) repoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "repository",
		Short:   "Communicate with JCR Repository",
		Aliases: []string{"repo"},
	}
	cmd.AddCommand(c.repoNodeCmd())

	return cmd
}

func (c *CLI) repoNodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "CRUD operations on JCR repository",
	}
	cmd.AddCommand(c.repoNodeReadCmd())
	cmd.AddCommand(c.repoNodeSaveCmd())
	cmd.AddCommand(c.repoNodeDeleteCmd())
	cmd.AddCommand(c.repoNodeCopyCmd())
	cmd.AddCommand(c.repoNodeMoveCmd())
	cmd.AddCommand(c.repoNodeChildrenCmd())
	cmd.AddCommand(c.repoNodeDownloadCmd())

	return cmd
}

func (c *CLI) repoNodeReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "read",
		Short:   "Read node",
		Aliases: []string{"get"},
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}
			node := repoNodeByFlags(cmd, *instance)
			c.SetOutput("node", node)
			c.Ok("node read")
		},
	}
	repoNodeDefineFlags(cmd)
	return cmd
}

func (c *CLI) repoNodeChildrenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "children",
		Short:   "Read node children",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}
			node := repoNodeByFlags(cmd, *instance)
			children, err := node.Children()
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("children", pkg.NewRepoNodeList(children))
			c.Ok("node children read")
		},
	}
	repoNodeDefineFlags(cmd)
	return cmd
}

func (c *CLI) repoNodeSaveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "save",
		Short:   "Create or update node",
		Aliases: []string{"create", "update"},
		Run: func(cmd *cobra.Command, args []string) {
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			var props map[string]any
			err = c.ReadInput(&props)
			if err != nil {
				c.Fail(fmt.Sprintf("cannot save node as input props cannot be parsed: %s", err))
				return
			}
			saved, err := pkg.InstanceProcess(c.aem, instances, func(instance pkg.Instance) (map[string]any, error) {
				node := repoNodeByFlags(cmd, instance)
				changed, err := node.SaveWithChanged(props)
				if err != nil {
					return nil, err
				}
				return map[string]any{
					OutputChanged: changed,
					"node":        node,
					"instance":    instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("saved", saved)
			if mapsx.SomeHas(saved, OutputChanged, true) {
				c.Changed("node saved")
			} else {
				c.Ok("node already saved (up-to-date)")
			}
		},
	}
	repoNodeDefineFlags(cmd)
	return cmd
}

func (c *CLI) repoNodeDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete node",
		Aliases: []string{"del", "remove", "rm"},
		Run: func(cmd *cobra.Command, args []string) {
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			deleted, err := pkg.InstanceProcess(c.aem, instances, func(instance pkg.Instance) (map[string]any, error) {
				node := repoNodeByFlags(cmd, instance)
				changed, err := node.DeleteWithChanged()
				if err != nil {
					return nil, err
				}
				return map[string]any{
					OutputChanged: changed,
					"node":        node,
					"instance":    instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("deleted", deleted)
			if mapsx.SomeHas(deleted, OutputChanged, true) {
				c.Changed("node deleted")
			} else {
				c.Ok("node already deleted (does not exist)")
			}
		},
	}
	repoNodeDefineFlags(cmd)
	return cmd
}

func (c *CLI) repoNodeCopyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "copy",
		Short:   "Copy node",
		Aliases: []string{"cp"},
		Run: func(cmd *cobra.Command, args []string) {
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			sourcePath, _ := cmd.Flags().GetString("source-path")
			targetPath, _ := cmd.Flags().GetString("target-path")
			copied, err := pkg.InstanceProcess(c.aem, instances, func(instance pkg.Instance) (map[string]any, error) {
				sourceNode := instance.Repo().Node(sourcePath)
				changed, err := sourceNode.CopyWithChanged(targetPath)
				if err != nil {
					return nil, err
				}
				targetNode := instance.Repo().Node(targetPath)
				return map[string]any{
					OutputChanged: changed,
					"sourceNode":  sourceNode,
					"targetNode":  targetNode,
					"instance":    instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("copied", copied)
			if mapsx.SomeHas(copied, OutputChanged, true) {
				c.Changed("node copied")
			} else {
				c.Ok("node already copied (up-to-date)")
			}
		},
	}
	cmd.Flags().StringP("source-path", "s", "", "Source path")
	cmd.MarkFlagRequired("source-path")
	cmd.Flags().StringP("target-path", "t", "", "Target path")
	cmd.MarkFlagRequired("target-path")
	return cmd
}

func (c *CLI) repoNodeMoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "move",
		Short:   "Move node",
		Aliases: []string{"mv"},
		Run: func(cmd *cobra.Command, args []string) {
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			sourcePath, _ := cmd.Flags().GetString("source-path")
			targetPath, _ := cmd.Flags().GetString("target-path")
			replace, _ := cmd.Flags().GetBool("replace")
			moved, err := pkg.InstanceProcess(c.aem, instances, func(instance pkg.Instance) (map[string]any, error) {
				sourceNode := instance.Repo().Node(sourcePath)
				changed, err := sourceNode.MoveWithChanged(targetPath, replace)
				if err != nil {
					return nil, err
				}
				targetNode := instance.Repo().Node(targetPath)
				return map[string]any{
					OutputChanged: changed,
					"sourceNode":  sourceNode,
					"targetNode":  targetNode,
					"instance":    instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("moved", moved)
			if mapsx.SomeHas(moved, OutputChanged, true) {
				c.Changed("node moved")
			} else {
				c.Ok("node already moved (up-to-date)")
			}
		},
	}
	cmd.Flags().StringP("source-path", "s", "", "Source path")
	cmd.MarkFlagRequired("source-path")
	cmd.Flags().StringP("target-path", "t", "", "Target path")
	cmd.MarkFlagRequired("target-path")
	cmd.Flags().BoolP("replace", "r", false, "Replace target node if it already exists")
	return cmd
}

func (c *CLI) repoNodeDownloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "download",
		Short:   "Download node pointing to file",
		Aliases: []string{"dl"},
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}
			node := repoNodeByFlags(cmd, *instance)
			targetFile, _ := cmd.Flags().GetString("target-file")
			force, _ := cmd.Flags().GetBool("force")
			changed := false
			if force {
				err = node.Download(targetFile)
				changed = true
			} else {
				changed, err = node.DownloadWithChanged(targetFile)
			}
			if err != nil {
				c.Error(err)
				return
			}
			if changed {
				c.Changed("node downloaded")
			} else {
				c.Ok("node not downloaded (up-to-date)")
			}
			c.SetOutput("node", node)
			c.SetOutput("instance", instance)
			c.SetOutput("file", targetFile)
			c.Ok("node downloaded")
		},
	}
	repoNodeDefineFlags(cmd)
	cmd.Flags().StringP("target-file", "t", "", "Target file path")
	cmd.Flags().BoolP("force", "f", false, "Download even when already downloaded")
	return cmd
}

func repoNodeDefineFlags(cmd *cobra.Command) {
	cmd.Flags().String("path", "", "Path")
	_ = cmd.MarkFlagRequired("path")
}

func repoNodeByFlags(cmd *cobra.Command, instance pkg.Instance) *pkg.RepoNode {
	path, _ := cmd.Flags().GetString("path")
	node := instance.Repo().Node(path)
	return &node
}
