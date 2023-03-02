package main

import (
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg"
	"github.com/wttech/aemc/pkg/common"
	"github.com/wttech/aemc/pkg/common/mapsx"
)

func (c *CLI) cryptoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "crypto",
		Short: "Manages Crypto Support",
	}
	cmd.AddCommand(c.cryptoSetupCmd())
	cmd.AddCommand(c.cryptoProtectCmd())
	return cmd
}

func (c *CLI) cryptoSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "setup",
		Aliases: []string{"configure"},
		Short:   "Setup Crypto keys",
		Run: func(cmd *cobra.Command, args []string) {
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}

			hmacFile, _ := cmd.Flags().GetString("hmac-file")
			masterFile, _ := cmd.Flags().GetString("master-file")

			configured, err := pkg.InstanceProcess(c.aem, instances, func(instance pkg.Instance) (map[string]any, error) {
				changed, err := instance.Crypto().Setup(hmacFile, masterFile)
				if err != nil {
					return nil, err
				}
				return map[string]any{
					OutputChanged: changed,
					"instance":    instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("configured", configured)
			if mapsx.SomeHas(configured, OutputChanged, true) {
				if err := c.aem.InstanceManager().AwaitStarted(instances); err != nil {
					c.Error(err)
					return
				}
				c.Changed("Crypto configured")
			} else {
				c.Ok("Crypto already configured")
			}
		},
	}
	cmd.Flags().StringP("hmac-file", "h", common.LibDir+"/crypto/hmac", "Path to file 'hmac'")
	cmd.MarkFlagRequired("hmac-file")
	cmd.Flags().StringP("master-file", "m", common.LibDir+"/crypto/master", "Path to file 'master'")
	cmd.MarkFlagRequired("master-file")
	return cmd
}

func (c *CLI) cryptoProtectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "protect",
		Aliases: []string{"encrypt"},
		Short:   "Protect value",
		Run: func(cmd *cobra.Command, args []string) {
			instance, err := c.aem.InstanceManager().One()
			if err != nil {
				c.Error(err)
				return
			}
			plainValue, _ := cmd.Flags().GetString("value")
			unprotectedValue, err := instance.Crypto().Protect(plainValue)
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("value", unprotectedValue)
			c.Ok("value protected")
		},
	}
	cmd.Flags().StringP("value", "v", "", "Value to protect")
	cmd.MarkFlagRequired("value")
	return cmd
}
