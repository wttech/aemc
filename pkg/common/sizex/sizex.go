package sizex

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
)

// Parse reads a human size like '100GB' or '512 MiB'; an empty value gives 0.
func Parse(value string) (uint64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	size, err := humanize.ParseBytes(value)
	if err != nil {
		return 0, fmt.Errorf("invalid size '%s': %w", value, err)
	}
	return size, nil
}

// ParseThreshold reads a size given either as a percentage of a total (e.g. '10%') or as an absolute size (e.g. '50GB').
// An empty value gives 0.
func ParseThreshold(value string, total uint64) (uint64, error) {
	value = strings.TrimSpace(value)
	percentText, isPercent := strings.CutSuffix(value, "%")
	if !isPercent {
		return Parse(value)
	}
	percent, err := strconv.ParseFloat(strings.TrimSpace(percentText), 64)
	if err != nil || percent < 0 || percent > 100 {
		return 0, fmt.Errorf("invalid percentage '%s': expected a number from 0 to 100", value)
	}
	return uint64(float64(total) * percent / 100), nil
}
