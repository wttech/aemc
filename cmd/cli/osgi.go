package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg"
	"github.com/wttech/aemc/pkg/common/mapsx"
)

func (c *CLI) osgiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "osgi",
		Short: "Communicate with OSGi Framework",
	}
	cmd.AddCommand(c.osgiBundleCmd())
	cmd.AddCommand(c.osgiConfigCmd())
	return cmd
}

func (c *CLI) osgiBundleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bundle",
		Short: "Manage OSGi bundles",
	}
	cmd.AddCommand(c.osgiBundleInstall())
	cmd.AddCommand(c.osgiBundleUninstall())
	cmd.AddCommand(c.osgiBundleListCmd())
	cmd.AddCommand(c.osgiBundleReadCmd())
	cmd.AddCommand(c.osgiBundleStartCmd())
	cmd.AddCommand(c.osgiBundleStopCmd())
	cmd.AddCommand(c.osgiBundleRestartCmd())
	return cmd
}

func (c *CLI) osgiBundleInstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install OSGi bundle(s)",
		Run: func(cmd *cobra.Command, args []string) {
			path, _ := cmd.Flags().GetString("file")
			installed, err := c.aem.InstanceManager().ProcessAll(func(instance pkg.Instance) (map[string]any, error) {
				changed, err := instance.OSGI().BundleManager().InstallWithChanged(path)
				if err != nil {
					return nil, err
				}
				bundle, err := instance.OSGI().BundleManager().ByFile(path)
				if err != nil {
					return nil, err
				}
				if changed {
					err = bundle.AwaitStarted()
					if err != nil {
						return nil, err
					}
				}
				return map[string]any{
					"changed":  changed,
					"bundle":   bundle,
					"instance": instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("installed", installed)
			if mapsx.HasSome(installed, "changed", true) {
				c.Changed("bundle installed")
			} else {
				c.Ok("bundle already installed")
			}
		},
	}
	osgiBundleDefineFileFlag(cmd)
	return cmd
}

func (c *CLI) osgiBundleUninstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall OSGi bundle(s)",
		Run: func(cmd *cobra.Command, args []string) {
			uninstalled, err := c.aem.InstanceManager().ProcessAll(func(instance pkg.Instance) (map[string]any, error) {
				bundle, err := osgiBundleByFlags(cmd, instance)
				if err != nil {
					return nil, err
				}
				changed, err := bundle.UninstallWithChanged()
				if err != nil {
					return nil, err
				}
				return map[string]any{
					"changed":  changed,
					"instance": instance,
					"bundle":   bundle,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("uninstalled", uninstalled)
			if mapsx.HasSome(uninstalled, "changed", true) {
				c.Changed("bundle uninstalled")
			} else {
				c.Ok("bundle already uninstalled")
			}
		},
	}
	osgiBundleDefineFlags(cmd)
	return cmd
}

func (c *CLI) osgiBundleListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List OSGi bundles",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			bundles, err := instance.OSGI().BundleManager().List()
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("bundles", bundles)
			c.Ok("bundles listed")
		},
	}
	return cmd
}

func (c *CLI) osgiBundleReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "read",
		Short:   "Read OSGi bundle details",
		Aliases: []string{"get", "find"},
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			bundle, err := osgiBundleByFlags(cmd, *instance)
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("bundle", bundle)
			c.Ok("bundle read")
		},
	}
	osgiBundleDefineFlags(cmd)
	return cmd
}

func (c *CLI) osgiBundleStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start OSGi bundle",
		Run: func(cmd *cobra.Command, args []string) {
			started, err := c.aem.InstanceManager().ProcessAll(func(instance pkg.Instance) (map[string]any, error) {
				bundle, err := osgiBundleByFlags(cmd, instance)
				if err != nil {
					return nil, err
				}
				changed, err := bundle.StartWithChanged()
				if err != nil {
					return nil, err
				}
				if changed {
					err = bundle.AwaitStarted()
					if err != nil {
						return nil, err
					}
				}
				return map[string]any{
					"changed":  changed,
					"instance": instance,
					"bundle":   bundle,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("started", started)
			if mapsx.HasSome(started, "changed", true) {
				c.Changed("bundle started")
			} else {
				c.Ok("bundle already started (up-to-date)")
			}
		},
	}
	osgiBundleDefineFlags(cmd)
	return cmd
}

func (c *CLI) osgiBundleStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop OSGi bundle",
		Run: func(cmd *cobra.Command, args []string) {
			stopped, err := c.aem.InstanceManager().ProcessAll(func(instance pkg.Instance) (map[string]any, error) {
				bundle, err := osgiBundleByFlags(cmd, instance)
				if err != nil {
					return nil, err
				}
				changed, err := bundle.StopWithChanged()
				if err != nil {
					return nil, err
				}
				if changed {
					err = bundle.AwaitStopped()
					if err != nil {
						c.Error(err)
					}
				}
				return map[string]any{
					"changed":  changed,
					"bundle":   bundle,
					"instance": instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("stopped", stopped)
			if mapsx.HasSome(stopped, "changed", true) {
				c.Changed("bundle stopped")
			} else {
				c.Ok("bundle already stopped (up-to-date)")
			}
		},
	}
	osgiBundleDefineFlags(cmd)
	return cmd
}

func (c *CLI) osgiBundleRestartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart OSGi bundle",
		Run: func(cmd *cobra.Command, args []string) {
			restarted, err := c.aem.InstanceManager().ProcessAll(func(instance pkg.Instance) (map[string]any, error) {
				bundle, err := osgiBundleByFlags(cmd, instance)
				if err != nil {
					return nil, err
				}
				err = bundle.Restart()
				if err != nil {
					return nil, err
				}
				err = bundle.AwaitStarted()
				if err != nil {
					c.Error(err)
				}
				return map[string]any{
					"changed":  true,
					"bundle":   bundle,
					"instance": instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("restarted", restarted)
			c.Changed("bundle restarted")
		},
	}
	osgiBundleDefineFlags(cmd)
	return cmd
}

func osgiBundleDefineFileFlag(cmd *cobra.Command) {
	cmd.Flags().String("file", "", "Local bundle JAR file")
	_ = cmd.MarkFlagRequired("file")
}

func osgiBundleDefineFlags(cmd *cobra.Command) {
	cmd.Flags().String("symbolic-name", "", "Symbolic Name")
	cmd.Flags().String("file", "", "Local bundle JAR file")
	cmd.MarkFlagsMutuallyExclusive("symbolic-name", "file")
}

func osgiBundleByFlags(cmd *cobra.Command, i pkg.Instance) (*pkg.OSGiBundle, error) {
	symbolicName, _ := cmd.Flags().GetString("symbolic-name")
	if len(symbolicName) > 0 {
		bundle := i.OSGI().BundleManager().New(symbolicName)
		return &bundle, nil
	}
	file, _ := cmd.Flags().GetString("file")
	if len(file) > 0 {
		bundle, err := i.OSGI().BundleManager().ByFile(file)
		return bundle, err
	}
	return nil, fmt.Errorf("flag 'symbolic-name' or 'file' are required")
}

func (c *CLI) osgiConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config",
		Aliases: []string{"cfg"},
		Short:   "Manage OSGi configuration",
	}
	cmd.AddCommand(c.osgiConfigList())
	cmd.AddCommand(c.osgiConfigRead())
	cmd.AddCommand(c.osgiConfigSave())
	cmd.AddCommand(c.osgiConfigDelete())
	return cmd
}

func (c *CLI) osgiConfigList() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List OSGi configurations",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			configs, err := instance.OSGI().ConfigManager().FindAll()
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("configs", configs)
			c.Ok("configs listed")
		},
	}
	return cmd
}

func (c *CLI) osgiConfigRead() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "read",
		Short:   "Read OSGi configuration values",
		Aliases: []string{"get"},
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			config := osgiConfigFromFlag(cmd, *instance)
			c.SetOutput("config", config)
			c.Ok("config read")
		},
	}
	osgiConfigDefineFlags(cmd)
	return cmd
}

func (c *CLI) osgiConfigSave() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "save",
		Short:   "Save OSGi configuration values",
		Aliases: []string{"set"},
		Run: func(cmd *cobra.Command, args []string) {
			var props map[string]any
			err := c.ReadInput(&props)
			if err != nil {
				c.Fail(fmt.Sprintf("cannot save config as input props cannot be parsed: %s", err))
				return
			}
			saved, err := c.aem.InstanceManager().ProcessAll(func(instance pkg.Instance) (map[string]any, error) {
				config := osgiConfigFromFlag(cmd, instance)
				changed, err := config.SaveWithChanged(props)
				if err != nil {
					return nil, err
				}
				if changed {
					c.aem.InstanceManager().AwaitStartedOne(instance)
				}
				return map[string]any{
					"changed":  changed,
					"config":   config,
					"instance": instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("saved", saved)
			if mapsx.HasSome(saved, "changed", true) {
				c.Changed("config saved")
			} else {
				c.Ok("config already saved (up-to-date)")
			}
		},
	}
	osgiConfigDefineFlags(cmd)
	return cmd
}

func (c *CLI) osgiConfigDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete OSGi configuration",
		Aliases: []string{"del", "remove", "unset"},
		Run: func(cmd *cobra.Command, args []string) {
			deleted, err := c.aem.InstanceManager().ProcessAll(func(instance pkg.Instance) (map[string]any, error) {
				config := osgiConfigFromFlag(cmd, instance)
				changed, err := config.DeleteWithChanged()
				if err != nil {
					return nil, err
				}
				if changed {
					c.aem.InstanceManager().AwaitStartedOne(instance)
				}
				return map[string]any{
					"changed":  changed,
					"config":   config,
					"instance": instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("deleted", deleted)
			if mapsx.HasSome(deleted, "changed", true) {
				c.Changed("config deleted")
			} else {
				c.Ok("config already deleted (does not exist)")
			}
		},
	}
	osgiConfigDefineFlags(cmd)
	return cmd
}

func osgiConfigDefineFlags(cmd *cobra.Command) {
	cmd.Flags().String("pid", "", "PID")
	_ = cmd.MarkFlagRequired("pid")
}

func osgiConfigFromFlag(cmd *cobra.Command, i pkg.Instance) *pkg.OSGiConfig {
	pid, _ := cmd.Flags().GetString("pid")
	config := i.OSGI().ConfigManager().ByPID(pid)
	return &config
}
