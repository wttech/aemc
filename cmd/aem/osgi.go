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
	cmd.AddCommand(c.osgiComponentCmd())
	cmd.AddCommand(c.osgiConfigCmd())

	cmd.AddCommand(c.osgiRestartCmd())
	return cmd
}

func (c *CLI) osgiBundleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bundle",
		Aliases: []string{"bnd"},
		Short:   "Manage OSGi bundles",
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
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			installed, err := pkg.InstanceProcess(c.aem, instances, func(instance pkg.Instance) (map[string]any, error) {
				changed, err := instance.OSGI().BundleManager().InstallWithChanged(path)
				if err != nil {
					return nil, err
				}
				bundle, err := instance.OSGI().BundleManager().ByFile(path)
				if err != nil {
					return nil, err
				}
				if changed {
					if err := bundle.AwaitStarted(); err != nil {
						return nil, err
					}
				}
				return map[string]any{
					OutputChanged: changed,
					"bundle":      bundle,
					"instance":    instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("installed", installed)
			if mapsx.HasSome(installed, OutputChanged, true) {
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
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			uninstalled, err := pkg.InstanceProcess(c.aem, instances, func(instance pkg.Instance) (map[string]any, error) {
				bundle, err := osgiBundleByFlags(cmd, instance)
				if err != nil {
					return nil, err
				}
				changed, err := bundle.UninstallWithChanged()
				if err != nil {
					return nil, err
				}
				return map[string]any{
					OutputChanged: changed,
					"instance":    instance,
					"bundle":      bundle,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("uninstalled", uninstalled)
			if mapsx.HasSome(uninstalled, OutputChanged, true) {
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
			instance, err := c.aem.InstanceManager().One()
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
			instance, err := c.aem.InstanceManager().One()
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
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			started, err := pkg.InstanceProcess(c.aem, instances, func(instance pkg.Instance) (map[string]any, error) {
				bundle, err := osgiBundleByFlags(cmd, instance)
				if err != nil {
					return nil, err
				}
				changed, err := bundle.StartWithChanged()
				if err != nil {
					return nil, err
				}
				if changed {
					if err = bundle.AwaitStarted(); err != nil {
						return nil, err
					}
				}
				return map[string]any{
					OutputChanged: changed,
					"instance":    instance,
					"bundle":      bundle,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("started", started)
			if mapsx.HasSome(started, OutputChanged, true) {
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
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			stopped, err := pkg.InstanceProcess(c.aem, instances, func(instance pkg.Instance) (map[string]any, error) {
				bundle, err := osgiBundleByFlags(cmd, instance)
				if err != nil {
					return nil, err
				}
				changed, err := bundle.StopWithChanged()
				if err != nil {
					return nil, err
				}
				if changed {
					if err = bundle.AwaitStopped(); err != nil {
						c.Error(err)
					}
				}
				return map[string]any{
					OutputChanged: changed,
					"bundle":      bundle,
					"instance":    instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("stopped", stopped)
			if mapsx.HasSome(stopped, OutputChanged, true) {
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
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			restarted, err := pkg.InstanceProcess(c.aem, instances, func(instance pkg.Instance) (map[string]any, error) {
				bundle, err := osgiBundleByFlags(cmd, instance)
				if err != nil {
					return nil, err
				}
				if err = bundle.Restart(); err != nil {
					return nil, err
				}
				if err = bundle.AwaitStarted(); err != nil {
					c.Error(err)
				}
				return map[string]any{
					OutputChanged: true,
					"bundle":      bundle,
					"instance":    instance,
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

func (c *CLI) osgiComponentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "component",
		Aliases: []string{"cmp"},
		Short:   "Manage OSGi components",
	}
	cmd.AddCommand(c.osgiComponentListCmd())
	cmd.AddCommand(c.osgiComponentReadCmd())
	cmd.AddCommand(c.osgiComponentEnableCmd())
	cmd.AddCommand(c.osgiComponentDisableCmd())
	cmd.AddCommand(c.osgiComponentReenableCmd())
	return cmd
}

func (c *CLI) osgiComponentListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List OSGi components",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}
			bundles, err := instance.OSGI().ComponentManager().List()
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("components", bundles)
			c.Ok("components listed")
		},
	}
	return cmd
}

func (c *CLI) osgiComponentReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "read",
		Short:   "Read OSGi component details",
		Aliases: []string{"get", "find"},
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}
			cmp := osgiComponentFromFlag(cmd, *instance)
			c.SetOutput("component", cmp)
			c.Ok("component read")
		},
	}
	osgiComponentDefineFlags(cmd)
	return cmd
}

func (c *CLI) osgiComponentEnableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable OSGi component",
		Run: func(cmd *cobra.Command, args []string) {
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			enabled, err := pkg.InstanceProcess(c.aem, instances, func(instance pkg.Instance) (map[string]any, error) {
				cmp := osgiComponentFromFlag(cmd, instance)
				changed, err := cmp.EnableWithChanged()
				if err != nil {
					return nil, err
				}
				if changed {
					if err := cmp.AwaitEnabled(); err != nil {
						return nil, err
					}
				}
				return map[string]any{
					OutputChanged: changed,
					"instance":    instance,
					"component":   cmp,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("enabled", enabled)
			if mapsx.HasSome(enabled, OutputChanged, true) {
				c.Changed("component enabled")
			} else {
				c.Ok("component already enabled (up-to-date)")
			}
		},
	}
	osgiComponentDefineFlags(cmd)
	return cmd
}

func (c *CLI) osgiComponentDisableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable OSGi component",
		Run: func(cmd *cobra.Command, args []string) {
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			disabled, err := pkg.InstanceProcess(c.aem, instances, func(instance pkg.Instance) (map[string]any, error) {
				cmp := osgiComponentFromFlag(cmd, instance)
				changed, err := cmp.DisableWithChanged()
				if err != nil {
					return nil, err
				}
				if changed {
					if err := cmp.AwaitDisabled(); err != nil {
						return nil, err
					}
				}
				return map[string]any{
					OutputChanged: changed,
					"instance":    instance,
					"component":   cmp,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("disabled", disabled)
			if mapsx.HasSome(disabled, OutputChanged, true) {
				c.Changed("component disabled")
			} else {
				c.Ok("component already disabled (up-to-date)")
			}
		},
	}
	osgiComponentDefineFlags(cmd)
	return cmd
}

func (c *CLI) osgiComponentReenableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reenable",
		Short: "Reenable OSGi component",
		Run: func(cmd *cobra.Command, args []string) {
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			reenabled, err := pkg.InstanceProcess(c.aem, instances, func(instance pkg.Instance) (map[string]any, error) {
				cmp := osgiComponentFromFlag(cmd, instance)
				if err := cmp.Reenable(); err != nil {
					return nil, err
				}
				if err := cmp.AwaitEnabled(); err != nil {
					c.Error(err)
				}
				return map[string]any{
					OutputChanged: true,
					"component":   cmp,
					"instance":    instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("reenabled", reenabled)
			c.Changed("component reenabled")
		},
	}
	osgiComponentDefineFlags(cmd)
	return cmd
}

func osgiComponentDefineFlags(cmd *cobra.Command) {
	cmd.Flags().String("pid", "", "PID")
	_ = cmd.MarkFlagRequired("pid")
}

func osgiComponentFromFlag(cmd *cobra.Command, i pkg.Instance) *pkg.OSGiComponent {
	pid, _ := cmd.Flags().GetString("pid")
	cmp := i.OSGI().ComponentManager().ByPID(pid)
	return &cmp
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
			instance, err := c.aem.InstanceManager().One()
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
			instance, err := c.aem.InstanceManager().One()
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
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			var props map[string]any
			if err := c.ReadInput(&props); err != nil {
				c.Fail(fmt.Sprintf("cannot save config as input props cannot be parsed: %s", err))
				return
			}
			saved, err := pkg.InstanceProcess(c.aem, instances, func(instance pkg.Instance) (map[string]any, error) {
				config := osgiConfigFromFlag(cmd, instance)
				changed, err := config.SaveWithChanged(props)
				if err != nil {
					return nil, err
				}
				if changed {
					c.aem.InstanceManager().AwaitStartedOne(instance)
				}
				return map[string]any{
					OutputChanged: changed,
					"config":      config,
					"instance":    instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("saved", saved)
			if mapsx.HasSome(saved, OutputChanged, true) {
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
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			deleted, err := pkg.InstanceProcess(c.aem, instances, func(instance pkg.Instance) (map[string]any, error) {
				config := osgiConfigFromFlag(cmd, instance)
				changed, err := config.DeleteWithChanged()
				if err != nil {
					return nil, err
				}
				if changed {
					c.aem.InstanceManager().AwaitStartedOne(instance)
				}
				return map[string]any{
					OutputChanged: changed,
					"config":      config,
					"instance":    instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("deleted", deleted)
			if mapsx.HasSome(deleted, OutputChanged, true) {
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

func (c *CLI) osgiRestartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart OSGi framework",
		Run: func(cmd *cobra.Command, args []string) {
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			restarted, err := pkg.InstanceProcess(c.aem, instances, func(instance pkg.Instance) (map[string]any, error) {
				if err := instance.OSGI().Restart(); err != nil {
					return nil, err
				}
				return map[string]any{
					OutputChanged: true,
					"instance":    instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			if err := c.aem.InstanceManager().Await(instances); err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("restarted", restarted)
			c.Changed("framework restarted")
		},
	}
	osgiBundleDefineFlags(cmd)
	return cmd
}
