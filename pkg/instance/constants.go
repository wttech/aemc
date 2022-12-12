package instance

type ProcessingMode string

const (
	ProcessingAuto     = "auto"
	ProcessingParallel = "parallel"
	ProcessingSerial   = "serial"
)

func ProcessingModes() []string {
	return []string{ProcessingAuto, ProcessingParallel, ProcessingSerial}
}
