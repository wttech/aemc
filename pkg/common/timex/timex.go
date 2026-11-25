package timex

import (
	"time"
)

const fileTimestampLayout = "20060102150405"

func FileTimestamp(value time.Time) string {
	return value.Format(fileTimestampLayout)
}

func FileTimestampForNow() string {
	return FileTimestamp(time.Now())
}

// ParseFileTimestamp reads a timestamp written by FileTimestamp (in local time).
func ParseFileTimestamp(value string) (time.Time, error) {
	return time.ParseInLocation(fileTimestampLayout, value, time.Local)
}

func Human(time time.Time) string {
	return time.Format("2006-01-02 15:04:05")
}
