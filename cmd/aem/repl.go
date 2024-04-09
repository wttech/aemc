package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg"
	"github.com/wttech/aemc/pkg/replication"
)

func (c *CLI) replCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "replication",
		Short:   "Replicate nodes and manage agents",
		Aliases: []string{"repl"},
	}
	cmd.AddCommand(c.replAgentCmd())
	cmd.AddCommand(c.replActivateCmd())
	cmd.AddCommand(c.replDeactivateCmd())
	cmd.AddCommand(c.replActivateTreeCmd())

	return cmd
}

func (c *CLI) replActivateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "activate",
		Short:   "Activate single node",
		Aliases: []string{"act", "replicate", "repl"},
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}

			path := replActivationPathByFlags(cmd)
			if err := instance.Replication().Activate(path); err != nil {
				c.Error(err)
				return
			}

			c.SetOutput("path", path)
			c.Ok("path activated")
		},
	}
	replActivationFlags(cmd)

	return cmd
}

func (c *CLI) replDeactivateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "deactivate",
		Short:   "Deactivate single node",
		Aliases: []string{"deact", "unreplicate", "unrepl"},
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}

			path := replActivationPathByFlags(cmd)
			if err := instance.Replication().Deactivate(path); err != nil {
				c.Error(err)
				return
			}

			c.SetOutput("path", path)
			c.Ok("path deactivated")
		},
	}
	replActivationFlags(cmd)
	return cmd
}

func (c *CLI) replActivateTreeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tree",
		Short:   "Activate node and all its children",
		Aliases: []string{"activate-tree", "tree-activate"},
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}

			startPath, _ := cmd.Flags().GetString("path")
			onlyModified, _ := cmd.Flags().GetBool("only-modified")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			onlyActivated, _ := cmd.Flags().GetBool("only-activated")
			ignoreDeactivated, _ := cmd.Flags().GetBool("ignore-deactivated")
			opts := replication.ActivateTreeOpts{
				StartPath:         startPath,
				OnlyModified:      onlyModified,
				DryRun:            dryRun,
				OnlyActivated:     onlyActivated,
				IgnoreDeactivated: ignoreDeactivated,
			}

			if err := instance.Replication().ActivateTree(opts); err != nil {
				c.Error(err)
				return
			}

			c.SetOutput("path", startPath)
			c.Ok("tree activated")
		},
	}
	cmd.Flags().StringP("path", "p", "", "Path to node")
	_ = cmd.MarkFlagRequired("path")
	cmd.Flags().Bool("dry-run", false, "Only simulate activation")
	cmd.Flags().Bool("only-modified", false, "Only activate modified nodes")
	cmd.Flags().Bool("only-activated", false, "Only activate nodes that are not activated")
	cmd.Flags().Bool("ignore-deactivated", false, "Ignore deactivated nodes")

	return cmd
}

func replActivationFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("path", "p", "", "Path to node")
	_ = cmd.MarkFlagRequired("path")
}

func replActivationPathByFlags(cmd *cobra.Command) string {
	path, _ := cmd.Flags().GetString("path")
	return path
}

func (c *CLI) replAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "agent",
		Short:   "Managing replication agents",
		Aliases: []string{"ag"},
	}
	cmd.AddCommand(c.replAgentReadCmd())
	cmd.AddCommand(c.replAgentSetupCmd())
	cmd.AddCommand(c.replAgentDeleteCmd())

	return cmd
}

func (c *CLI) replAgentReadCmd() *cobra.Command {
	result := &cobra.Command{
		Use:     "read",
		Short:   "Read replication agent details",
		Aliases: []string{"get"},
		Run: func(cmd *cobra.Command, args []string) {
			i, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}
			replAgent := replAgentByFlags(cmd, i)
			c.SetOutput("replAgent", replAgent)
			c.Ok("replication agent details read")
		},
	}
	replAgentDefineFlags(result)
	return result
}

func (c *CLI) replAgentDeleteCmd() *cobra.Command {
	result := &cobra.Command{
		Use:     "delete",
		Short:   "Delete replication agent",
		Aliases: []string{"remove", "rm"},
		Run: func(cmd *cobra.Command, args []string) {
			i, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}
			replAgent := replAgentByFlags(cmd, i)
			changed, err := replAgent.Delete()
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("replAgent", replAgent)
			if changed {
				c.Changed("replication agent deleted")
			} else {
				c.Ok("replication agent already deleted (does not exist)")
			}
		},
	}
	replAgentDefineFlags(result)
	return result
}

func (c *CLI) replAgentSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "setup",
		Short:   "Setup replication agent",
		Aliases: []string{"configure"},
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}
			var props map[string]any
			if err = c.ReadInput(&props); err != nil {
				c.Fail(fmt.Sprintf("cannot setup replication agent as input props cannot be parsed: %s", err))
				return
			}
			replAgent := replAgentByFlags(cmd, instance)
			changed, err := replAgent.Setup(props)
			if err != nil {
				c.Error(err)
				return
			}
			if changed {
				if err = instance.Replication().Bundle().Restart(); err != nil {
					c.Error(err)
					return
				}
				if err := c.aem.InstanceManager().AwaitStartedOne(*instance); err != nil {
					c.Error(err)
					return
				}
			}
			c.SetOutput("replAgent", replAgent)
			if changed {
				c.Changed("replication agent set up")
			} else {
				c.Ok("replication agent already set up (up-to-date)")
			}
		},
	}
	replAgentDefineFlags(cmd)

	return cmd
}

func replAgentDefineFlags(cmd *cobra.Command) {
	cmd.Flags().String("location", "", "Location")
	_ = cmd.MarkFlagRequired("location")
	cmd.Flags().String("name", "", "Name")
	_ = cmd.MarkFlagRequired("name")
}

func replAgentByFlags(cmd *cobra.Command, i *pkg.Instance) *pkg.ReplAgent {
	location, _ := cmd.Flags().GetString("location")
	name, _ := cmd.Flags().GetString("name")
	replAgent := i.Repo().ReplAgent(location, name)
	return &replAgent
}
