package main

import (
	"fmt"
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
	cmd.AddCommand(c.oakIndexReindexAllCmd())
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

func (c *CLI) oakIndexReindexAllCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reindex-all",
		Short: "Reindex all OAK indexes",
		Run: func(cmd *cobra.Command, args []string) {
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			reindexed, err := pkg.InstanceProcess(c.aem, instances, func(instance pkg.Instance) (map[string]any, error) {
				indexes, err := instance.OAK().IndexManager().ReindexAll()
				if err != nil {
					return nil, err
				}

				return map[string]any{
					OutputChanged: true,
					"instance":    instance,
					"indexes":     indexes,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("reindexed", reindexed)
			if mapsx.SomeHas(reindexed, OutputChanged, true) {
				c.Changed("indexes reindexed")
			} else {
				c.Ok("indexes already reindexed (up-to-date)")
			}
		},
	}
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
