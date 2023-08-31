package main

import (
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg"
	"github.com/wttech/aemc/pkg/common/mapsx"
)

func (c *CLI) sslCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssl",
		Short: "Manages SSL (HTTPS) Support",
	}
	cmd.AddCommand(c.sslSetupCmd())
	return cmd
}

func (c *CLI) sslSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "setup",
		Aliases: []string{"configure"},
		Short:   "Setup SSL",
		Run: func(cmd *cobra.Command, args []string) {
			instances, err := c.aem.InstanceManager().Some()
			if err != nil {
				c.Error(err)
				return
			}

			keyStorePassword, _ := cmd.Flags().GetString("keystore-password")
			trustStorePassword, _ := cmd.Flags().GetString("truststore-password")
			certificateFile, _ := cmd.Flags().GetString("certificate-file")
			privateKeyFile, _ := cmd.Flags().GetString("private-key-file")
			httpsHostname, _ := cmd.Flags().GetString("https-hostname")
			httpsPort, _ := cmd.Flags().GetString("https-port")

			configured, err := pkg.InstanceProcess(c.aem, instances, func(instance pkg.Instance) (map[string]any, error) {
				changed, err := instance.SSL().Setup(keyStorePassword, trustStorePassword, certificateFile, privateKeyFile, httpsHostname, httpsPort)
				if err != nil {
					return nil, err
				}
				return map[string]any{
					OutputChanged:  changed,
					OutputInstance: instance,
				}, nil
			})
			if err != nil {
				c.Error(err)
				return
			}
			c.SetOutput("configured", configured)

			if mapsx.SomeHas(configured, OutputChanged, true) {
				c.Changed("SSL set up")
			} else {
				c.Ok("SSL already set up (up-to-date)")
			}
		},
	}
	cmd.Flags().String("keystore-password", "", "Keystore password")
	cmd.Flags().String("truststore-password", "", "Truststore password")
	cmd.Flags().String("certificate-file", "", "Certificate file (PEM format)")
	cmd.Flags().String("private-key-file", "", "Private key file (DER or PEM format)")
	cmd.Flags().String("https-hostname", "", "HTTPS hostname")
	cmd.Flags().String("https-port", "", "HTTPS port")
	_ = cmd.MarkFlagRequired("keystore-password")
	_ = cmd.MarkFlagRequired("truststore-password")
	_ = cmd.MarkFlagRequired("certificate-file")
	_ = cmd.MarkFlagRequired("private-key-file")
	_ = cmd.MarkFlagRequired("https-hostname")
	_ = cmd.MarkFlagRequired("https-port")
	return cmd
}
