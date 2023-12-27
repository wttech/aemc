package instance

import (
	"github.com/samber/lo"
	"github.com/wttech/aemc/pkg/common/stringsx"
	"strings"
)

func MatchSome(id string, value string, patterns []string) bool {
	return lo.SomeBy(patterns, func(p string) bool { return Match(id, value, p) })
}

func Match(id string, value string, pattern string) bool {
	if !strings.Contains(pattern, ":") {
		return stringsx.Match(value, pattern)
	}

	parts := strings.Split(pattern, ":")

	instancePattern := parts[0]
	instancePatterns := strings.Split(instancePattern, ",")
	if !stringsx.MatchSome(id, instancePatterns) {
		return false
	}

	specificPattern := parts[1]
	specificPatterns := strings.Split(specificPattern, ",")
	return stringsx.MatchSome(value, specificPatterns)
}
