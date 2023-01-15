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
