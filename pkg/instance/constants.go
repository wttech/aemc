package instance

import (
	_ "embed"
)

const (
	IDDelimiter          = "_"
	URLLocalAuthor       = "http://localhost:4502"
	URLLocalPublish      = "http://localhost:4503"
	PasswordDefault      = "admin"
	UserDefault          = "admin"
	LocationLocal        = "local"
	LocationRemote       = "remote"
	RoleAuthorPortSuffix = "02"
	ClassifierDefault    = ""
	AemVersionUnknown    = "unknown"
)

type ProcessingMode string

const (
	ProcessingAuto     = "auto"
	ProcessingParallel = "parallel"
	ProcessingSerial   = "serial"
)

func ProcessingModes() []string {
	return []string{ProcessingAuto, ProcessingParallel, ProcessingSerial}
}

//go:embed resource/cbp.exe
var CbpExecutable []byte

//go:embed resource/oak-run/set-password.groovy
var OakRunSetPassword string

type Role string

const (
	RoleAuthor  Role = "author"
	RolePublish Role = "publish"
)
