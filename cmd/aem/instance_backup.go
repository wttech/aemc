package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (c *CLI) instanceBackupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "backup",
		Aliases: []string{"bak"},
		Short:   "Manages AEM instance backups",
	}
	cmd.AddCommand(c.instanceBackupListCmd())
	cmd.AddCommand(c.instanceBackupMakeCmd())
	cmd.AddCommand(c.instanceBackupUseCmd())
	cmd.AddCommand(c.instanceBackupPerformCmd())
	cmd.AddCommand(c.instanceBackupRestoreCmd())
	return cmd
}

func (c *CLI) instanceBackupListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "Lists available AEM instance backups",
		Run: func(cmd *cobra.Command, args []string) {
			backups, err := c.aem.InstanceManager().ListBackups()
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("backups", backups)
			c.Ok("instance backup listed")
		},
	}
	return cmd
}

func (c *CLI) instanceBackupMakeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "make",
		Aliases: []string{"mk", "create"},
		Short:   "Makes single AEM instance backup",
		Run: func(cmd *cobra.Command, args []string) {
			instanceManager := c.aem.InstanceManager()
			localInstance, err := instanceManager.OneLocal()
			if err != nil {
				c.Error(err)
				return
			}
			instance := localInstance.Instance()

			file, _ := cmd.Flags().GetString("file")
			if file == "" {
				file = localInstance.ProposeBackupFileToMake()
			}

			running := localInstance.IsRunning()
			if running {
				if localInstance.StopAndAwait(); err != nil {
					c.Error(err)
					return
				}
			}
			if err := localInstance.MakeBackup(file); err != nil {
				c.Error(err)
				return
			}
			if running {
				if err := localInstance.StartAndAwait(); err != nil {
					c.Error(err)
					return
				}
			}

			c.SetOutput("instance", instance)
			c.SetOutput("file", file)
			c.Ok("instance backup made")
		},
	}
	cmd.Flags().String("file", "", "Local file path")
	return cmd
}

func (c *CLI) instanceBackupUseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "use",
		Aliases: []string{"read", "load"},
		Short:   "Uses single AEM instance backup",
		Run: func(cmd *cobra.Command, args []string) {
			instanceManager := c.aem.InstanceManager()
			localInstance, err := instanceManager.OneLocal()
			if err != nil {
				c.Error(err)
				return
			}
			deleteCreated, _ := cmd.Flags().GetBool("delete-created")
			if !deleteCreated && localInstance.IsCreated() {
				c.Fail("instance already created; to delete it use flag '--delete-created'")
				return
			}
			instance := localInstance.Instance()
			file, _ := cmd.Flags().GetString("file")
			if file == "" {
				fileProposal, err := localInstance.ProposeBackupFileToUse()
				if err != nil {
					c.Error(err)
					return
				}
				file = fileProposal
			}
			running := localInstance.IsRunning()
			if running {
				if localInstance.StopAndAwait(); err != nil {
					c.Error(err)
					return
				}
			}
			if err := localInstance.UseBackup(file, deleteCreated); err != nil {
				c.Error(err)
				return
			}
			if running {
				if err := localInstance.StartAndAwait(); err != nil {
					c.Error(err)
					return
				}
			}
			c.SetOutput("instance", instance)
			c.SetOutput("file", file)
			c.Ok("instance backup used")
		},
	}
	cmd.Flags().String("file", "", "Local file path")
	cmd.Flags().Bool("delete-created", false, "Delete already created instance")
	return cmd
}

func (c *CLI) instanceBackupPerformCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "perform",
		Short: "Performs backups of created AEM instances",
		Run: func(cmd *cobra.Command, args []string) {
			instanceManager := c.aem.InstanceManager()
			instances, err := instanceManager.SomeLocals()
			if err != nil {
				c.Error(err)
				return
			}
			performed := []map[string]any{}
			for _, instance := range instances {
				local := instance.Local()
				if !local.IsCreated() {
					log.Infof("%s > skipping making backup as it is not created", instance.ID())
					continue
				}
				file := local.ProposeBackupFileToMake()
				running := local.IsRunning()
				if running {
					if local.StopAndAwait(); err != nil {
						c.Error(err)
						return
					}
				}
				if err := local.MakeBackup(file); err != nil {
					c.Error(err)
					return
				}
				if running {
					if err := local.StartAndAwait(); err != nil {
						c.Error(err)
						return
					}
				}
				performed = append(performed, map[string]any{
					"instance": instance,
					"file":     file,
				})
			}
			c.SetOutput("performed", performed)
			if len(performed) > 0 {
				c.Changed("instances backups performed")
			} else {
				c.Ok("no instance backups performed")
			}
		},
	}
	return cmd
}

func (c *CLI) instanceBackupRestoreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restores from backups uncreated AEM instances",
		Run: func(cmd *cobra.Command, args []string) {
			instanceManager := c.aem.InstanceManager()
			instances, err := instanceManager.SomeLocals()
			if err != nil {
				c.Error(err)
				return
			}
			restored := []map[string]any{}
			for _, instance := range instances {
				local := instance.Local()
				if local.IsCreated() {
					log.Infof("%s > skipping using backup as it is already created", instance.ID())
					continue
				}
				file, err := local.ProposeBackupFileToUse()
				if err != nil {
					log.Warnf("%s > skipping using backup as cannot find any: %s", instance.ID(), err)
					continue
				}
				running := local.IsRunning()
				if running {
					if local.StopAndAwait(); err != nil {
						c.Error(err)
						return
					}
				}
				if err := local.UseBackup(file, false); err != nil {
					c.Error(err)
					return
				}
				if running {
					if err := local.StartAndAwait(); err != nil {
						c.Error(err)
						return
					}
				}
				restored = append(restored, map[string]any{
					"instance": instance,
					"file":     file,
				})
			}
			c.SetOutput("restored", restored)
			if len(restored) > 0 {
				c.Changed("instances restored from backups")
			} else {
				c.Ok("no instances restored from backups")
			}
		},
	}
	return cmd
}
