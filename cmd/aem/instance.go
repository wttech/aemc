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

	cmd.AddCommand(&cobra.Command{
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
	})
	cmd.AddCommand(&cobra.Command{
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
	})
	cmd.AddCommand(&cobra.Command{
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
	})
	cmd.AddCommand(&cobra.Command{
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
	})
	cmd.AddCommand(&cobra.Command{
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
	})
	cmd.AddCommand(&cobra.Command{
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
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Checks status of AEM instance(s)",
		Run: func(cmd *cobra.Command, args []string) {
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("instances", instances)
			c.Ok("instance(s) status returned")
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "await",
		Aliases: []string{"wait"},
		Short:   "Awaits stable AEM instance(s)",
		Run: func(cmd *cobra.Command, args []string) {
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			c.aem.InstanceManager().Await(instances)
			c.SetOutput("instances", instances)
			c.Ok("instance(s) start awaited")
		},
	})

	return cmd
}
