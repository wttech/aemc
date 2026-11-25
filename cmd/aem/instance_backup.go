package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg"
	"github.com/wttech/aemc/pkg/common/sizex"
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
	cmd.AddCommand(c.instanceBackupPruneCmd())
	return cmd
}

func (c *CLI) instanceBackupListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "Lists available AEM instance backups",
		Run: func(cmd *cobra.Command, args []string) {
			backups, err := c.aem.InstanceManager().Backup.List()
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
			opts, err := c.instanceBackupMakeOpts(cmd)
			if err != nil {
				c.Error(err)
				return
			}
			opts.File, _ = cmd.Flags().GetString("file")
			file, err := instanceManager.Backup.Make(localInstance, opts)
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("instance", localInstance.Instance())
			c.SetOutput("file", file)
			c.Ok("instance backup made")
		},
	}
	cmd.Flags().String("file", "", "Local file path")
	instanceBackupMakeFlags(cmd)
	return cmd
}

func instanceBackupMakeFlags(cmd *cobra.Command) {
	cmd.Flags().String("name", "", "Backup name describing the captured state (e.g. 'clean with prereqs'); kept apart from other names when pruning")
	cmd.Flags().Bool("skip-space-check", false, "Skip checking that enough disk space stays free after making backup")
}

func (c *CLI) instanceBackupMakeOpts(cmd *cobra.Command) (pkg.BackupMakeOpts, error) {
	name, err := c.instanceBackupName(cmd)
	if err != nil {
		return pkg.BackupMakeOpts{}, err
	}
	skipSpaceCheck, _ := cmd.Flags().GetBool("skip-space-check")
	return pkg.BackupMakeOpts{Name: name, SkipSpaceCheck: skipSpaceCheck}, nil
}

func (c *CLI) instanceBackupName(cmd *cobra.Command) (string, error) {
	name, _ := cmd.Flags().GetString("name")
	return c.aem.InstanceManager().Backup.NormalizeName(name)
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
			name, err := c.instanceBackupName(cmd)
			if err != nil {
				c.Error(err)
				return
			}
			file, _ := cmd.Flags().GetString("file")
			file, err = instanceManager.Backup.Use(localInstance, pkg.BackupUseOpts{File: file, Name: name, DeleteCreated: deleteCreated})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("instance", localInstance.Instance())
			c.SetOutput("file", file)
			c.Ok("instance backup used")
		},
	}
	cmd.Flags().String("file", "", "Local file path")
	cmd.Flags().String("name", "", "Use the newest backup of this name")
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
			opts, err := c.instanceBackupMakeOpts(cmd)
			if err != nil {
				c.Error(err)
				return
			}
			performed := []map[string]any{}
			for _, instance := range instances {
				local := instance.Local()
				if !local.IsCreated() {
					log.Infof("%s > skipping making backup as it is not created", instance.IDColor())
					continue
				}
				file, err := instanceManager.Backup.Make(local, opts)
				if err != nil {
					c.Error(err)
					return
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
	instanceBackupMakeFlags(cmd)
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
			name, err := c.instanceBackupName(cmd)
			if err != nil {
				c.Error(err)
				return
			}
			restored := []map[string]any{}
			for _, instance := range instances {
				local := instance.Local()
				if local.IsCreated() {
					log.Infof("%s > skipping using backup as it is already created", instance.IDColor())
					continue
				}
				if _, err := instanceManager.Backup.Newest(local.Name(), name, ""); err != nil {
					log.Warnf("%s > skipping using backup as cannot find any: %s", instance.IDColor(), err)
					continue
				}
				file, err := instanceManager.Backup.Use(local, pkg.BackupUseOpts{Name: name})
				if err != nil {
					c.Error(err)
					return
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
	cmd.Flags().String("name", "", "Use the newest backups of this name")
	return cmd
}

func (c *CLI) instanceBackupPruneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "prune",
		Aliases: []string{"rotate"},
		Short:   "Deletes old AEM instance backups according to the retention policy",
		Run: func(cmd *cobra.Command, args []string) {
			instanceManager := c.aem.InstanceManager()
			policy, err := instanceManager.LocalOpts.Backup.PrunePolicy()
			if err != nil {
				c.Error(err)
				return
			}
			if cmd.Flags().Changed("keep-last") {
				policy.KeepLast, _ = cmd.Flags().GetInt("keep-last")
			}
			if cmd.Flags().Changed("max-size") {
				maxSize, _ := cmd.Flags().GetString("max-size")
				if policy.MaxSize, err = sizex.Parse(maxSize); err != nil {
					c.Error(err)
					return
				}
			}
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			result, err := instanceManager.Backup.PruneWith(policy, dryRun, "")
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("pruned", result)
			if dryRun || len(result.Deleted) == 0 {
				c.Ok("no instance backups pruned")
			} else {
				c.Changed("instance backups pruned")
			}
		},
	}
	cmd.Flags().Int("keep-last", 0, "Newest backups kept per instance and backup name; 0 keeps all (default from 'instance.local.backup.prune.keep_last')")
	cmd.Flags().String("max-size", "", "Cap on the total size of backups, e.g. '100GB'; empty means no cap (default from 'instance.local.backup.prune.max_size')")
	cmd.Flags().Bool("dry-run", false, "Only print which backups would be deleted")
	return cmd
}
