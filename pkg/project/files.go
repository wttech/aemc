package project

//go:generate go run ./wrapper

import (
	"embed"
)

// Files of each project kind, shared by all runners.

//go:embed instance
var InstanceFiles embed.FS

//go:embed app_classic
var AppClassicFiles embed.FS

//go:embed app_cloud
var AppCloudFiles embed.FS

// RunnerFiles are the files of each runner: 'runner/<runner>/common' for every project kind, then 'runner/<runner>/<kind>'.
//
//go:embed runner
var RunnerFiles embed.FS
