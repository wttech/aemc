package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/common"
	"os"
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
			scriptFiles, err := os.ReadDir(common.ScriptDir)
			if err != nil {
				c.Error(fmt.Errorf("cannot list scripts in dir '%s'", common.ScriptDir))
			}
			var scripts []string
			for _, file := range scriptFiles {
				if strings.HasSuffix(file.Name(), ".sh") {
					scripts = append(scripts, strings.TrimSuffix(file.Name(), ".sh"))
				}
			}
			c.SetOutput("gettingStarted", fmt.Sprintf(strings.Join([]string{
				"The next step is providing AEM files (JAR or SDK ZIP, license, service packs) to directory '" + common.LibDir + "'.",
				"Alternatively, instruct the tool where these files are located by adjusting properties: 'dist_file', 'license_file' in configuration file '" + cfg.FileDefault + "'.",
				"To avoid problems with IDE performance, make sure to exclude from indexing the directory '" + common.HomeDir + "'.",
				"Finally, use control scripts to manage AEM instances:",
				"",

				"sh aemw [%s]",

				"",
				"It is also possible to run individual AEM Compose CLI commands separately.",
				"Discover available commands by running:",
				"",

				"sh aemw --help",
			}, "\n"), strings.Join(scripts, "|")))
			c.Ok("initialized properly")
		},
	}
	return cmd
}
