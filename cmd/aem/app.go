package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var appVersion = "<unknown>"
var appCommit = "<unknown>"
var appCommitDate = "<unknown>"

type AppInfo struct {
	Version    string `yaml:"version" json:"version"`
	Commit     string `yaml:"commit" json:"commit"`
	CommitDate string `yaml:"commit_date" json:"commitDate"`
}

func NewAppInfo() AppInfo {
	return AppInfo{
		Version:    appVersion,
		Commit:     appCommit,
		CommitDate: appCommitDate,
	}
}

func (a AppInfo) String() string {
	return fmt.Sprintf("AEM CLI %s (commit %s on %s)", a.Version, a.Commit, a.CommitDate)
}

func (c *CLI) versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print app details including version",
		Run: func(cmd *cobra.Command, args []string) {
			c.SetOutput("app", NewAppInfo())
			c.Ok("application details printed")
		},
	}
}
