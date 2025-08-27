package sdk

const (
	OSAuto    = "auto"
	OSUnix    = "unix"
	OSWindows = "windows"
)

func OsTypes() []string {
	return []string{OSAuto, OSUnix, OSWindows}
}
