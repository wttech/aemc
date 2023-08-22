package mapsx

import (
	"github.com/google/go-cmp/cmp"
	"github.com/samber/lo"
	"golang.org/x/exp/maps"
	"strings"
)

func Equal(current map[string]any, updated map[string]any) bool {
	return EqualIgnoring(current, updated, []string{})
}

func EqualIgnoring(current map[string]any, updated map[string]any, ignored []string) bool {
	before := maps.Clone(current)
	predicted := maps.Clone(current)
	maps.Copy(predicted, updated)
	for _, name := range ignored {
		delete(predicted, name)
		delete(before, name)
	}
	return cmp.Equal(before, predicted)
}

func Has[T comparable](data map[string]any, key string, value T) bool {
	actualValue, ok := data[key]
	if ok {
		return value == actualValue.(T)
	}
	return false
}

func SomeHas[T comparable](dataList []map[string]any, key string, value T) bool {
	return lo.SomeBy(dataList, func(data map[string]any) bool { return Has(data, key, value) })
}

// FromString is a workaround for https://github.com/spf13/pflag/issues/129
func FromString(value string) map[string]string {
	result := map[string]string{}
	if value == "" {
		return result
	}
	for _, str := range strings.Split(value, ",") {
		parts := strings.Split(str, "=")
		if len(parts) == 2 {
			k := parts[0]
			v := parts[1]
			result[k] = v
		}
	}
	return result
}
