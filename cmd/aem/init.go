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
			scripts, err := c.initFindScriptNames()
			if err != nil {
				c.Error(err)
				return
			}

			//projectType := cmd.Flags().GetString("project-type")

			// TODO
			// unpack cloud/classic project files

			c.SetOutput("gettingStarted", fmt.Sprintf(strings.Join([]string{
				"The next step is providing AEM files (JAR or SDK ZIP, license, service packs) to directory '" + common.LibDir + "'.",
				"Alternatively, instruct the tool where these files are located by adjusting properties: 'dist_file', 'license_file' in configuration file '" + cfg.FileDefault + "'.",
				"Make sure to exclude the directory '" + common.HomeDir + "'from VCS versioning and IDE indexing.",
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
	cmd.Flags().String("project-type", "cloud", "Project type (cloud|classic)")
	return cmd
}

func (c *CLI) initFindScriptNames() ([]string, error) {
	scriptFiles, err := os.ReadDir(common.ScriptDir)
	if err != nil {
		return nil, fmt.Errorf("cannot list scripts in dir '%s'", common.ScriptDir)
	}
	var scripts []string
	for _, file := range scriptFiles {
		if strings.HasSuffix(file.Name(), ".sh") {
			scripts = append(scripts, strings.TrimSuffix(file.Name(), ".sh"))
		}
	}
	return scripts, nil
}
