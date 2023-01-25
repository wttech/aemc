package instance

import (
	_ "embed"
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

//go:embed resource/oak-run-1.42.0.jar
var OakRunJar []byte

//go:embed resource/oak-run_set-password.groovy
var OakRunSetPassword string
