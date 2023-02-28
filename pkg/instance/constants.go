package instance

import (
	_ "embed"
)

const (
	IDDelimiter          = "_"
	URLLocalAuthor       = "http://127.0.0.1:4502"
	URLLocalPublish      = "http://127.0.0.1:4503"
	PasswordDefault      = "admin"
	UserDefault          = "admin"
	LocationLocal        = "local"
	LocationRemote       = "remote"
	RoleAuthorPortSuffix = "02"
	ClassifierDefault    = ""
	AemVersionUnknown    = "<unknown>"
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

// CbpExecutable is a recompiled binary from code at 'https://ritchielawrence.github.io/cmdow' to avoid false-positive antivirus detection
//
//go:embed resource/cbpow.exe
var CbpExecutable []byte

//go:embed resource/oak-run/set-password.groovy
var OakRunSetPassword string

type Role string

const (
	CbpExecutableFilename = "cbpow.exe"

	RoleAuthor  Role = "author"
	RolePublish Role = "publish"
)
