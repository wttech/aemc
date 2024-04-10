package main

import (
	"github.com/sirupsen/logrus"
	"regexp"
)

type LogCustomFormatter struct {
	*logrus.TextFormatter
}

func (f *LogCustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	bytes, err := f.TextFormatter.Format(entry)
	if err != nil {
		return nil, err
	}
	message := string(bytes)
	message = logInstanceIDRemoveDuplicate(message)
	return []byte(message), nil
}

var logInstanceIDRegex = regexp.MustCompile(`(\x1b\[\d+m)?([\w_]+)(\x1b\[0m)? > `)

func logInstanceIDRemoveDuplicate(message string) string {
	seen := make(map[string]bool)
	message = logInstanceIDRegex.ReplaceAllStringFunc(message, func(match string) string {
		id := logInstanceIDRegex.FindStringSubmatch(match)[2]
		if seen[id] {
			return ""
		}
		seen[id] = true
		return match
	})
	return message
}
