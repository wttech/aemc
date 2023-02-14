package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wttech/aemc/pkg/cfg"
	"github.com/wttech/aemc/pkg/common"
	"github.com/wttech/aemc/pkg/project"
	"strings"
)

func (c *CLI) initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "initialize",
		Aliases: []string{"init"},
		Short:   "Initializes project files and configuration",
		Run: func(cmd *cobra.Command, args []string) {
			kindName, _ := cmd.Flags().GetString("project-kind")
			kind, err := c.project.DetermineKind(kindName)
			if err != nil {
				c.Error(err)
				return
			}
			changed, err := c.project.InitializeWithChanged(kind)
			if err != nil {
				c.Error(err)
				return
			}
			scripts, err := c.project.FindScriptNames()
			if err != nil {
				c.Error(err)
				return
			}
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
			if changed {
				c.Changed("project initialized")
			} else {
				c.Ok("project already initialized")
			}
		},
	}
	cmd.Flags().String("project-kind", project.KindClassic, fmt.Sprintf("Project kind (%s)", strings.Join(project.KindStrings(), "|")))
	return cmd
}
