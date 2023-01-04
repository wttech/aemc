package stringsx

import (
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"
)

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func Percent(num, total, decimals int) string {
	value := 0.0

	if total != 0 {
		value = float64(num) / float64(total) * float64(100)
	}

	return fmt.Sprintf("%."+strconv.Itoa(decimals)+"f%%", value)
}

func PercentExplained(num, total, decimals int) string {
	return fmt.Sprintf("%d/%d=%s", num, total, Percent(num, total, decimals))
}

func MatchPattern(value, pattern string) bool {
	matched, _ := path.Match(pattern, value)
	return matched
}

func MatchAnyPattern(value string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := path.Match(pattern, value); matched {
			return true
		}
	}
	return false
}

func Between(value string, a string, b string) string {
	posFirst := strings.Index(value, a)
	if posFirst == -1 {
		return ""
	}
	posLast := strings.Index(value, b)
	if posLast == -1 {
		return ""
	}
	posFirstAdjusted := posFirst + len(a)
	if posFirstAdjusted >= posLast {
		return ""
	}
	return value[posFirstAdjusted:posLast]
}

func BeforeLast(value string, a string) string {
	pos := strings.LastIndex(value, a)
	if pos == -1 {
		return value
	}
	return value[0:pos]
}

func AfterLast(value string, a string) string {
	pos := strings.LastIndex(value, a)
	if pos == -1 {
		return value
	}
	adjustedPos := pos + len(a)
	if adjustedPos >= len(value) {
		return ""
	}
	return value[adjustedPos:]
}

var snakeMatchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var snakeMatchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func SnakeCase(str string) string {
	snake := snakeMatchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = snakeMatchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
