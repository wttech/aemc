package mapsx

import (
	"github.com/google/go-cmp/cmp"
	"github.com/samber/lo"
	"golang.org/x/exp/maps"
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

func HasSome[T comparable](dataList []map[string]any, key string, value T) bool {
	return lo.SomeBy(dataList, func(data map[string]any) bool { return Has(data, key, value) })
}
