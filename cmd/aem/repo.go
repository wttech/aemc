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
	cmd.AddCommand(c.repoNodeRead())
	cmd.AddCommand(c.repoNodeSave())
	cmd.AddCommand(c.repoNodeDelete())

	return cmd
}

func (c *CLI) repoNodeRead() *cobra.Command {
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

func (c *CLI) repoNodeSave() *cobra.Command {
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
					"changed":  changed,
					"node":     node,
					"instance": instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("saved", saved)
			if mapsx.HasSome(saved, "changed", true) {
				c.Changed("node saved")
			} else {
				c.Ok("node already saved (up-to-date)")
			}
		},
	}
	repoNodeDefineFlags(cmd)
	return cmd
}

func (c *CLI) repoNodeDelete() *cobra.Command {
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
					"changed":  changed,
					"node":     node,
					"instance": instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("deleted", deleted)
			if mapsx.HasSome(deleted, "changed", true) {
				c.Changed("node deleted")
			} else {
				c.Ok("node already deleted (does not exist)")
			}
		},
	}
	repoNodeDefineFlags(cmd)
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
