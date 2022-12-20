package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg"
)

func (c *CLI) replCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "replication",
		Short:   "Manage replication (agents)",
		Aliases: []string{"repl"},
	}
	cmd.AddCommand(c.replAgentCmd())

	return cmd
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
				c.aem.InstanceManager().AwaitStartedOne(*instance) // TODO if changed restart bundle afterwards (togglable)
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
