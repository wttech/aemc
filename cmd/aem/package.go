package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg"
	"github.com/wttech/aemc/pkg/common/httpx"
	"github.com/wttech/aemc/pkg/common/mapsx"
	"github.com/wttech/aemc/pkg/common/pathx"
	"strings"
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
	cmd.AddCommand(c.pkgPurgeCmd())
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
			instance, err := c.aem.InstanceManager().One()
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
			instance, err := c.aem.InstanceManager().One()
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
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			path, err := c.pkgPathByFlags(cmd)
			if err != nil {
				c.Error(err)
				return
			}
			uploaded, err := pkg.InstanceProcess(c.aem, instances, func(instance pkg.Instance) (map[string]any, error) {
				changed, err := instance.PackageManager().UploadWithChanged(path)
				if err != nil {
					return nil, err
				}
				p, err := instance.PackageManager().ByFile(path)
				if err != nil {
					return nil, err
				}
				return map[string]any{
					OutputChanged: changed,
					"package":     p,
					"instance":    instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			if err := c.aem.InstanceManager().AwaitStarted(InstancesChanged(uploaded)); err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("uploaded", uploaded)
			if mapsx.SomeHas(uploaded, OutputChanged, true) {
				c.Changed("package uploaded")
			} else {
				c.Ok("package not uploaded (up-to-date)")
			}
		},
	}
	pkgDefineFileAndUrlFlags(cmd)
	return cmd
}

func (c *CLI) pkgInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install package(s)",
		Run: func(cmd *cobra.Command, args []string) {
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			installed, err := pkg.InstanceProcess(c.aem, instances, func(instance pkg.Instance) (map[string]any, error) {
				p, err := pkgByFlags(cmd, instance)
				if err != nil {
					return nil, err
				}
				changed, err := p.InstallWithChanged()
				if err != nil {
					return nil, err
				}
				return map[string]any{
					OutputChanged: changed,
					"package":     p,
					"instance":    instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			if err := c.aem.InstanceManager().AwaitStarted(InstancesChanged(installed)); err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("installed", installed)
			if mapsx.SomeHas(installed, OutputChanged, true) {
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
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			path, err := c.pkgPathByFlags(cmd)
			if err != nil {
				c.Error(err)
				return
			}
			force, _ := cmd.Flags().GetBool("force")
			deployed, err := pkg.InstanceProcess(c.aem, instances, func(instance pkg.Instance) (map[string]any, error) {
				changed := false
				if force {
					err = instance.PackageManager().Deploy(path)
					changed = true
				} else {
					changed, err = instance.PackageManager().DeployWithChanged(path)
				}
				if err != nil {
					return nil, err
				}
				p, err := instance.PackageManager().ByFile(path)
				if err != nil {
					return nil, err
				}
				return map[string]any{
					OutputChanged: changed,
					"package":     p,
					"instance":    instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			if err := c.aem.InstanceManager().AwaitStarted(InstancesChanged(deployed)); err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("deployed", deployed)
			if mapsx.SomeHas(deployed, OutputChanged, true) {
				c.Changed("package deployed")
			} else {
				c.Ok("package already deployed (up-to-date)")
			}
		},
	}
	pkgDefineFileAndUrlFlags(cmd)
	cmd.Flags().BoolP("force", "f", false, "Deploy even when already deployed")
	return cmd
}

func (c *CLI) pkgUninstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall package",
		Run: func(cmd *cobra.Command, args []string) {
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			uninstalled, err := pkg.InstanceProcess(c.aem, instances, func(instance pkg.Instance) (map[string]any, error) {
				p, err := pkgByFlags(cmd, instance)
				if err != nil {
					return nil, err
				}
				changed, err := p.UninstallWithChanged()
				if err != nil {
					return nil, err
				}
				return map[string]any{
					OutputChanged: changed,
					"package":     p,
					"instance":    instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			if err := c.aem.InstanceManager().AwaitStarted(InstancesChanged(uninstalled)); err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("uninstalled", uninstalled)
			if mapsx.SomeHas(uninstalled, OutputChanged, true) {
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
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			deleted, err := pkg.InstanceProcess(c.aem, instances, func(instance pkg.Instance) (map[string]any, error) {
				p, err := pkgByFlags(cmd, instance)
				if err != nil {
					return nil, err
				}
				changed, err := p.DeleteWithChanged()
				if err != nil {
					return nil, err
				}
				return map[string]any{
					OutputChanged: changed,
					"package":     p,
					"instance":    instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			if err := c.aem.InstanceManager().AwaitStarted(InstancesChanged(deleted)); err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("deleted", deleted)
			if mapsx.SomeHas(deleted, OutputChanged, true) {
				c.Changed("package deleted")
			} else {
				c.Ok("package not deleted (does not exist)")
			}
		},
	}
	pkgDefineFlags(cmd)
	return cmd
}

func (c *CLI) pkgPurgeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "purge",
		Short: "Purge package (uninstall and delete)",
		Run: func(cmd *cobra.Command, args []string) {
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			purged, err := pkg.InstanceProcess(c.aem, instances, func(instance pkg.Instance) (map[string]any, error) {
				p, err := pkgByFlags(cmd, instance)
				if err != nil {
					return nil, err
				}
				state, err := p.State()
				if err != nil {
					return nil, err
				}
				changed := false
				if state.Exists {
					if state.Data.Installed() {
						uninstalled, err := p.UninstallWithChanged()
						if err != nil {
							return nil, err
						}
						if uninstalled {
							changed = true
						}
					}
					deleted, err := p.DeleteWithChanged()
					if err != nil {
						return nil, err
					}
					changed = changed || deleted
				}
				return map[string]any{
					OutputChanged: changed,
					"package":     p,
					"instance":    instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			if err := c.aem.InstanceManager().AwaitStarted(InstancesChanged(purged)); err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("purged", purged)
			if mapsx.SomeHas(purged, OutputChanged, true) {
				c.Changed("package purged")
			} else {
				c.Ok("package already purged")
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
			instance, err := c.aem.InstanceManager().One()
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
			if err := c.aem.InstanceManager().AwaitStartedOne(*instance); err != nil {
				c.Error(err)
				return
			}
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
		fileGlobbed, err := pathx.GlobSome(file)
		if err != nil {
			return nil, err
		}
		descriptor, err := instance.PackageManager().ByFile(fileGlobbed)
		return descriptor, err
	}
	path, _ := cmd.Flags().GetString("path")
	if len(path) > 0 {
		descriptor, err := instance.PackageManager().ByPath(path)
		return descriptor, err
	}
	return nil, fmt.Errorf("flag 'pid' or 'file' or 'path' are required")
}

func pkgDefineFileAndUrlFlags(cmd *cobra.Command) {
	cmd.Flags().String("file", "", "Local ZIP path")
	cmd.Flags().String("url", "", "URL to ZIP")
	cmd.MarkFlagsMutuallyExclusive("file", "url")
}

func (c *CLI) pkgPathByFlags(cmd *cobra.Command) (string, error) {
	url, _ := cmd.Flags().GetString("url")
	if len(url) > 0 {
		fileName := httpx.FileNameFromURL(url)
		if !strings.HasSuffix(fileName, ".zip") {
			return "", fmt.Errorf("package URL does not contain file name but it should '%s'", url)
		}
		path := c.aem.BaseOpts().TmpDir + "/" + fileName
		err := httpx.DownloadOnce(url, path)
		if err != nil {
			return "", err
		}
		return path, nil
	}
	file, _ := cmd.Flags().GetString("file")
	if len(file) > 0 {
		fileGlobbed, err := pathx.GlobSome(file)
		if err != nil {
			return "", err
		}
		return fileGlobbed, nil
	}
	return "", fmt.Errorf("flag 'file' or 'url' are required")
}
