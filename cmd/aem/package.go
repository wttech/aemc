package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg"
	"github.com/wttech/aemc/pkg/common/mapsx"
)

func (c *CLI) pkgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "package",
		Short:   "Communicates with Package Manager",
		Aliases: []string{"pkg"},
	}
	cmd.AddCommand(c.pkgListCmd())
	cmd.AddCommand(c.pkgUploadCmd())
	cmd.AddCommand(c.pkgInstallCmd())
	cmd.AddCommand(c.pkgDeployCmd())
	cmd.AddCommand(c.pkgUninstallCmd())
	cmd.AddCommand(c.pkgDeleteCmd())
	cmd.AddCommand(c.pkgBuildCmd())
	cmd.AddCommand(c.pkgFindCmd())
	return cmd
}

func (c *CLI) pkgFindCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "find",
		Short:   "Find package(s)",
		Aliases: []string{"get"},
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			p, err := pkgByFlags(cmd, *instance)
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("package", p)
			c.SetOutput("instance", instance)
			c.Ok("package found")
		},
	}
	pkgDefineFlags(cmd)
	return cmd
}

func (c *CLI) pkgListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List packages",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			pkgs, err := instance.PackageManager().List()
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("instance", instance)
			c.SetOutput("packages", pkgs)
			c.Ok("packages listed")
		},
	}
}

func (c *CLI) pkgUploadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload package(s)",
		Run: func(cmd *cobra.Command, args []string) {
			uploaded, err := c.aem.InstanceManager().ProcessAll(func(instance pkg.Instance) (map[string]any, error) {
				path := pkgPathByFlag(cmd)
				changed, err := instance.PackageManager().UploadWithChanged(path)
				if err != nil {
					return nil, err
				}
				if changed {
					c.aem.InstanceManager().AwaitStartedOne(instance)
				}
				p, err := instance.PackageManager().ByFile(path)
				if err != nil {
					return nil, err
				}
				return map[string]any{
					"changed":  changed,
					"package":  p,
					"instance": instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("uploaded", uploaded)
			if mapsx.HasSome(uploaded, "changed", true) {
				c.Changed("package uploaded")
			} else {
				c.Ok("package not uploaded (up-to-date)")
			}
		},
	}
	pkgDefineFileFlag(cmd)
	return cmd
}

func (c *CLI) pkgInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install package(s)",
		Run: func(cmd *cobra.Command, args []string) {
			changedAny := false
			installed, err := c.aem.InstanceManager().ProcessAll(func(instance pkg.Instance) (map[string]any, error) {
				p, err := pkgByFlags(cmd, instance)
				if err != nil {
					return nil, err
				}
				changed, err := p.InstallWithChanged()
				if err != nil {
					return nil, err
				}
				if changed {
					changedAny = true
					c.aem.InstanceManager().AwaitStartedOne(instance)
				}
				return map[string]any{
					"changed":  changed,
					"package":  p,
					"instance": instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("installed", installed)
			if changedAny {
				c.Changed("package installed")
			} else {
				c.Ok("package already installed (up-to-date)")
			}
		},
	}
	pkgDefineFlags(cmd)
	return cmd
}

func (c *CLI) pkgDeployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy package(s)",
		Run: func(cmd *cobra.Command, args []string) {
			deployed, err := c.aem.InstanceManager().ProcessAll(func(instance pkg.Instance) (map[string]any, error) {
				path := pkgPathByFlag(cmd)
				changed, err := instance.PackageManager().DeployWithChanged(path)
				if err != nil {
					return nil, err
				}
				if changed {
					c.aem.InstanceManager().AwaitStartedOne(instance)
				}
				p, err := instance.PackageManager().ByFile(path)
				if err != nil {
					return nil, err
				}
				return map[string]any{
					"changed":  changed,
					"package":  p,
					"instance": instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("deployed", deployed)
			if mapsx.HasSome(deployed, "changed", true) {
				c.Changed("package deployed")
			} else {
				c.Ok("package already deployed (up-to-date)")
			}
		},
	}
	pkgDefineFileFlag(cmd)
	return cmd
}

func (c *CLI) pkgUninstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall package",
		Run: func(cmd *cobra.Command, args []string) {
			uninstalled, err := c.aem.InstanceManager().ProcessAll(func(instance pkg.Instance) (map[string]any, error) {
				p, err := pkgByFlags(cmd, instance)
				if err != nil {
					return nil, err
				}
				changed, err := p.UninstallWithChanged()
				if err != nil {
					return nil, err
				}
				if changed {
					c.aem.InstanceManager().AwaitStartedOne(instance)
				}
				return map[string]any{
					"changed":  changed,
					"package":  p,
					"instance": instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("uninstalled", uninstalled)
			if mapsx.HasSome(uninstalled, "changed", true) {
				c.Changed("package uninstalled")
			} else {
				c.Ok("package already uninstalled")
			}
		},
	}
	pkgDefineFlags(cmd)
	return cmd
}

func (c *CLI) pkgDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete package",
		Run: func(cmd *cobra.Command, args []string) {
			deleted, err := c.aem.InstanceManager().ProcessAll(func(instance pkg.Instance) (map[string]any, error) {
				p, err := pkgByFlags(cmd, instance)
				if err != nil {
					return nil, err
				}
				changed, err := p.DeleteWithChanged()
				if err != nil {
					return nil, err
				}
				if changed {
					c.aem.InstanceManager().AwaitStartedOne(instance)
				}
				return map[string]any{
					"changed":  changed,
					"package":  p,
					"instance": instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("deleted", deleted)
			if mapsx.HasSome(deleted, "changed", true) {
				c.Changed("package deleted")
			} else {
				c.Ok("package not deleted (does not exist)")
			}
		},
	}
	pkgDefineFlags(cmd)
	return cmd
}

func (c *CLI) pkgBuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build package(s)",
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			p, err := pkgByFlags(cmd, *instance)
			if err != nil {
				c.Error(err)
				return
			}
			err = p.Build()
			if err != nil {
				c.Error(err)
				return
			}
			c.aem.InstanceManager().AwaitStartedOne(*instance)
			c.SetOutput("package", p)
			c.SetOutput("instance", instance)
			c.Changed("package built")
		},
	}
	pkgDefineFlags(cmd)
	return cmd
}

func pkgDefineFlags(cmd *cobra.Command) {
	cmd.Flags().String("pid", "", "ID (group:name:version)'")
	cmd.Flags().String("file", "", "Local path on file system")
	cmd.Flags().String("path", "", "Remote path on AEM repository")
	cmd.MarkFlagsMutuallyExclusive("pid", "file", "path")
}

func pkgByFlags(cmd *cobra.Command, instance pkg.Instance) (*pkg.Package, error) {
	pid, _ := cmd.Flags().GetString("pid")
	if len(pid) > 0 {
		desc, err := instance.PackageManager().ByPID(pid)
		return desc, err
	}
	file, _ := cmd.Flags().GetString("file")
	if len(file) > 0 {
		descriptor, err := instance.PackageManager().ByFile(file)
		return descriptor, err
	}
	path, _ := cmd.Flags().GetString("path")
	if len(path) > 0 {
		descriptor, err := instance.PackageManager().ByPath(path)
		return descriptor, err
	}
	return nil, fmt.Errorf("flag 'pid' or 'file' or 'path' are required")
}

func pkgDefineFileFlag(cmd *cobra.Command) {
	cmd.Flags().String("file", "", "Local ZIP path")
	_ = cmd.MarkFlagRequired("file")
}

func pkgPathByFlag(cmd *cobra.Command) string {
	path, _ := cmd.Flags().GetString("file")
	return path
}