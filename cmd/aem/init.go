package main

import (
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/common"
	"strings"
)

func (c *CLI) initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "initialize",
		Aliases: []string{"init"},
		Short:   "Initializes configuration and dependencies",
		Run: func(cmd *cobra.Command, args []string) {
			if !c.config.IsInitialized() {
				if err := c.config.Initialize(); err != nil {
					c.Error(err)
					return
				}
			}
			c.SetOutput("gettingStarted", strings.Join([]string{
				"The next step is providing AEM files (JAR or SDK ZIP, license, service packs) to directory '" + common.LibDir + "'.",
				"Alternatively, instruct the tool where these files are located by adjusting properties: 'dist_file', 'license_file' in configuration file '" + cfg.FileDefault + "'.",
				"To avoid problems with IDE performance, make sure to exclude from indexing the directory '" + common.HomeDir + "'.",
				"Finally, use control scripts to manage AEM instances:",
				"",

				"sh aemw [setup|resetup|up|down|restart]",

				"",
				"It is also possible to run individual AEM Compose CLI commands separately.",
				"Discover available commands by running:",
				"",

				"sh aemw --help",
			}, "\n"))
			c.Ok("initialized properly")
		},
	}
	return cmd
}
