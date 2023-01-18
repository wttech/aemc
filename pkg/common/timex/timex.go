package timex

import "time"

func FileTimestamp(value time.Time) string {
	return value.Format("20060102150405")
}

func FileTimestampForNow() string {
	return FileTimestamp(time.Now())
}
