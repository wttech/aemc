package main

import (
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg"
	"github.com/wttech/aemc/pkg/common/mapsx"
)

func (c *CLI) cryptoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "crypto",
		Short: "Manages Crypto Support",
	}
	cmd.AddCommand(c.cryptoConfigureCmd())
	cmd.AddCommand(c.cryptoProtectCmd())
	return cmd
}

func (c *CLI) cryptoConfigureCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configure",
		Aliases: []string{"setup"},
		Short:   "Configure Crypto keys",
		Run: func(cmd *cobra.Command, args []string) {
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}
			configured, err := pkg.InstanceProcess(c.aem, instances, func(instance pkg.Instance) (map[string]any, error) {
				// TODO read keys from flags or by convention 'aem/etc/crypto/hmac' and 'aem/etc/crypto/master'
				return map[string]any{
					OutputChanged: true,
					"instance":    instance,
				}, nil
			})
			// TODO restart OSGi framework afterward
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("configured", configured)
			if mapsx.HasSome(configured, OutputChanged, true) {
				c.Changed("Crypto configured")
			} else {
				c.Ok("Crypto already configured")
			}
		},
	}
	return cmd
}

func (c *CLI) cryptoProtectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "protect",
		Aliases: []string{"encrypt"},
		Short:   "Protect value",
		Run: func(cmd *cobra.Command, args []string) {
			plainValue := ""               // TODO implement this
			unprotectedValue := plainValue // TODO implement this
			c.SetOutput("value", unprotectedValue)
			c.Ok("value protected")
		},
	}
	return cmd
}
