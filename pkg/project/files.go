package project

import (
	"embed"
)

//go:embed common
var CommonFiles embed.FS

//go:embed instance
var InstanceFiles embed.FS

//go:embed app_classic
var AppClassicFiles embed.FS

//go:embed app_cloud
var AppCloudFiles embed.FS
