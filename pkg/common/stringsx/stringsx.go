package stringsx

import (
	"fmt"
	"github.com/gobwas/glob"
	"github.com/samber/lo"
	"regexp"
	"strconv"
	"strings"
)

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

func Match(value, pattern string) bool {
	return glob.MustCompile(pattern).Match(value)
}

func MatchSome(value string, patterns []string) bool {
	return lo.SomeBy(patterns, func(p string) bool { return Match(value, p) })
}

func Between(str string, start string, end string) (result string) {
	s := strings.Index(str, start)
	if s == -1 {
		return
	}
	s += len(start)
	e := strings.Index(str[s:], end)
	if e == -1 {
		return
	}
	return str[s : s+e]
}

func Before(value string, a string) string {
	pos := strings.Index(value, a)
	if pos == -1 {
		return value
	}
	return value[0:pos]
}

func After(value string, a string) string {
	pos := strings.Index(value, a)
	if pos == -1 {
		return value
	}
	adjustedPos := pos + len(a)
	if adjustedPos >= len(value) {
		return ""
	}
	return value[adjustedPos:]
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

func HumanCase(str string) string {
	return strings.ReplaceAll(SnakeCase(str), "_", " ")
}
