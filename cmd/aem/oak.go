package main

import (
	"fmt"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg"
	"github.com/wttech/aemc/pkg/common/mapsx"
)

func (c *CLI) oakCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oak",
		Short: "Manages OAK repository",
	}
	cmd.AddCommand(c.oakIndexCmd())
	return cmd
}

func (c *CLI) oakIndexCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "index",
		Aliases: []string{"idx"},
		Short:   "Manage OAK indexes",
	}
	cmd.AddCommand(c.oakIndexListCmd())
	cmd.AddCommand(c.oakIndexReadCmd())
	cmd.AddCommand(c.oakIndexReindexCmd())
	cmd.AddCommand(c.oakIndexReindexBatchCmd())
	return cmd
}

func (c *CLI) oakIndexListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List OAK indexes",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}
			indexes, err := instance.OAK().IndexManager().List()
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("indexes", indexes)
			c.Ok("indexes listed")
		},
	}
	return cmd
}

func (c *CLI) oakIndexReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "read",
		Short:   "Read OAK index details",
		Aliases: []string{"get", "find"},
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}
			bundle, err := oakIndexByFlags(cmd, *instance)
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("index", bundle)
			c.Ok("index read")
		},
	}
	oakIndexDefineFlags(cmd)
	return cmd
}

func (c *CLI) oakIndexReindexCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reindex",
		Short: "Reindex OAK index",
		Run: func(cmd *cobra.Command, args []string) {
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			reindexed, err := pkg.InstanceProcess(c.aem, instances, func(instance pkg.Instance) (map[string]any, error) {
				index, err := oakIndexByFlags(cmd, instance)
				if err != nil {
					return nil, err
				}
				changed, err := index.ReindexWithChanged()
				if err != nil {
					return nil, err
				}
				if changed {
					if err = index.AwaitNotReindexed(); err != nil {
						return nil, err
					}
				}
				return map[string]any{
					OutputChanged: changed,
					"instance":    instance,
					"index":       index,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("reindexed", reindexed)
			if mapsx.SomeHas(reindexed, OutputChanged, true) {
				c.Changed("index reindexed")
			} else {
				c.Ok("index already reindexed (in progress)")
			}
		},
	}
	oakIndexDefineFlags(cmd)
	return cmd
}

func (c *CLI) oakIndexReindexBatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reindex-batch",
		Short: "Reindex OAK indexes in batch",
		Run: func(cmd *cobra.Command, args []string) {
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}

			batchId, _ := cmd.Flags().GetString("batch-id")
			namePatterns, _ := cmd.Flags().GetStringSlice("name-pattern")
			force, _ := cmd.Flags().GetBool("force")

			reindexed, err := pkg.InstanceProcess(c.aem, instances, func(instance pkg.Instance) (map[string]any, error) {
				indexes, err := instance.OAK().IndexManager().FindByName(namePatterns)
				indexNames := lo.Map(indexes, func(i pkg.OAKIndex, _ int) any { return i.Name() })

				if force {
					if err := instance.OAK().IndexManager().ReindexBatch(indexes); err != nil {
						return nil, err
					}
					return map[string]any{
						OutputChanged: true,
						"instance":    instance,
						"indexNames":  indexNames,
					}, nil
				}

				changed, err := instance.OAK().IndexManager().ReindexBatchWithChanged(batchId, indexes)
				if err != nil {
					return nil, err
				}
				return map[string]any{
					OutputChanged: changed,
					"instance":    instance,
					"batchId":     batchId,
					"indexNames":  indexNames,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("reindexed", reindexed)
			if mapsx.SomeHas(reindexed, OutputChanged, true) {
				c.Changed("indexes batch reindexed")
			} else {
				c.Ok("indexes batch already reindexed (up-to-date)")
			}
		},
	}
	cmd.Flags().StringP("batch-id", "b", "all", "Batch ID")
	cmd.Flags().StringSliceP("name-pattern", "n", []string{"*"}, "Index name pattern(s)")
	cmd.Flags().BoolP("force", "f", false, "Reindex even if already indexed")
	return cmd
}

func oakIndexDefineFlags(cmd *cobra.Command) {
	cmd.Flags().String("name", "", "Name")
	_ = cmd.MarkFlagRequired("name")
}

func oakIndexByFlags(cmd *cobra.Command, i pkg.Instance) (*pkg.OAKIndex, error) {
	name, _ := cmd.Flags().GetString("name")
	if len(name) > 0 {
		index := i.OAK().IndexManager().New(name)
		return &index, nil
	}
	return nil, fmt.Errorf("flag 'name' is required")
}
