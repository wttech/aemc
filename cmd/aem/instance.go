package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg/common/intsx"
)

func (c *CLI) instanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "instance",
		Aliases: []string{"inst"},
		Short:   "Manages AEM instance(s)",
	}
	cmd.AddCommand(c.instanceCreateCmd())
	cmd.AddCommand(c.instanceStartCmd())
	cmd.AddCommand(c.instanceStopCmd())
	cmd.AddCommand(c.instanceRestartCmd())
	cmd.AddCommand(c.instanceKillCmd())
	cmd.AddCommand(c.instanceDeleteCmd())
	cmd.AddCommand(c.instanceListCmd())
	cmd.AddCommand(c.instanceAwaitCmd())
	return cmd
}

func (c *CLI) instanceCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "create",
		Short:   "Creates AEM instance(s)",
		Aliases: []string{"make"},
		Run: func(cmd *cobra.Command, args []string) {
			localInstances, err := c.aem.InstanceManager().SomeLocals()
			if err != nil {
				c.Error(err)
				return
			}
			createdInstances, err := c.aem.InstanceManager().Create(localInstances)
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("created", createdInstances)
			if len(createdInstances) > 0 {
				c.Changed(fmt.Sprintf("created instance(s) (%d)", len(createdInstances)))
			} else {
				c.Ok("no instance(s) to create")
			}
		},
	}
}

func (c *CLI) instanceStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "start",
		Aliases: []string{"up"},
		Short:   "Starts AEM instance(s)",
		Run: func(cmd *cobra.Command, args []string) {
			localInstances, err := c.aem.InstanceManager().SomeLocals()
			if err != nil {
				c.Error(err)
				return
			}
			startedInstances, err := c.aem.InstanceManager().Start(localInstances)
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("started", startedInstances)
			if len(startedInstances) > 0 {
				c.Changed(fmt.Sprintf("started instance(s) (%d)", len(startedInstances)))
			} else {
				c.Ok("no instance(s) to start")
			}
		},
	}
}

func (c *CLI) instanceStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "stop",
		Aliases: []string{"down"},
		Short:   "Stops AEM instance(s)",
		Run: func(cmd *cobra.Command, args []string) {
			localInstances, err := c.aem.InstanceManager().SomeLocals()
			if err != nil {
				c.Error(err)
				return
			}
			stoppedInstances, err := c.aem.InstanceManager().Stop(localInstances)
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("stopped", stoppedInstances)
			if len(stoppedInstances) > 0 {
				c.Changed(fmt.Sprintf("stopped instance(s) (%d)", len(stoppedInstances)))
			} else {
				c.Ok("no instance(s) to stop")
			}
		},
	}
}

func (c *CLI) instanceRestartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restart",
		Short: "Restarts AEM instance(s)",
		Run: func(cmd *cobra.Command, args []string) {
			localInstances, err := c.aem.InstanceManager().SomeLocals()
			if err != nil {
				c.Error(err)
				return
			}
			stoppedInstances, err := c.aem.InstanceManager().Stop(localInstances)
			if err != nil {
				c.Error(err)
				return
			}
			startedInstances, err := c.aem.InstanceManager().Start(localInstances)
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("restarted", localInstances)
			if len(stoppedInstances) > 0 || len(startedInstances) > 0 {
				c.Changed(fmt.Sprintf("restarted instance(s) (%d)", intsx.MaxOf(len(stoppedInstances), len(startedInstances))))
			} else {
				c.Ok("no instance(s) to restart")
			}
		},
	}
}

func (c *CLI) instanceKillCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "kill",
		Aliases: []string{"ko"},
		Short:   "Kills AEM instance(s)",
		Run: func(cmd *cobra.Command, args []string) {
			localInstances, err := c.aem.InstanceManager().SomeLocals()
			if err != nil {
				c.Error(err)
				return
			}
			killedInstances, err := c.aem.InstanceManager().Kill(localInstances)
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("killed", killedInstances)
			if len(killedInstances) > 0 {
				c.Changed(fmt.Sprintf("killed instance(s) (%d)", len(killedInstances)))
			} else {
				c.Ok("no instance(s) killed")
			}
		},
	}
}

func (c *CLI) instanceDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "delete",
		Aliases: []string{"destroy"},
		Short:   "Deletes AEM instance(s)",
		Run: func(cmd *cobra.Command, args []string) {
			localInstances, err := c.aem.InstanceManager().SomeLocals()
			if err != nil {
				c.Error(err)
				return
			}
			deletedInstances, err := c.aem.InstanceManager().Delete(localInstances)
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("deleted", deletedInstances)
			if len(deletedInstances) > 0 {
				c.Changed(fmt.Sprintf("deleted instance(s) (%d)", len(deletedInstances)))
			} else {
				c.Ok("no instance(s) to delete")
			}
		},
	}
}

func (c *CLI) instanceAwaitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "await",
		Aliases: []string{"wait"},
		Short:   "Awaits stable AEM instance(s)",
		Run: func(cmd *cobra.Command, args []string) {
			doneNever, _ := cmd.Flags().GetBool("done-never")
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			manager := c.aem.InstanceManager()
			manager.CheckOpts.DoneNever = doneNever
			manager.Await(instances)
			c.SetOutput("instances", instances)
			c.Ok("instance(s) awaited")
		},
	}
	cmd.Flags().IntVar(&(c.config.Values().Instance.Check.DoneThreshold),
		"done-threshold", c.config.Values().Instance.Check.DoneThreshold,
		"Number of successful checks indicating done")
	cmd.Flags().Bool("done-never", false, "Repeat checks endlessly")
	return cmd
}

func (c *CLI) instanceListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "status"},
		Short:   "Lists all AEM instance(s)",
		Run: func(cmd *cobra.Command, args []string) {
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("instances", instances)
			c.Ok("instance(s) listed")
		},
	}
}
