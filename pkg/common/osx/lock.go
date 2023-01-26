package osx

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/common/pathx"
)

type Lock[T comparable] struct {
	path         string
	dataProvider func() T
}

func NewLock[T comparable](path string, dataProvider func() T) Lock[T] {
	return Lock[T]{path, dataProvider}
}

func (l Lock[T]) Lock() error {
	err := fmtx.MarshalToFile(l.path, l.dataProvider())
	if err != nil {
		return fmt.Errorf("cannot save lock file '%s': %w", l.path, err)
	}
	return nil
}

func (l Lock[T]) Unlock() error {
	if err := pathx.DeleteIfExists(l.path); err != nil {
		return fmt.Errorf("cannot delete lock file '%s': %w", l.path, err)
	}
	return nil
}

func (l Lock[T]) IsLocked() bool {
	return pathx.Exists(l.path)
}

func (l Lock[T]) DataCurrent() T {
	return l.dataProvider()
}

func (l Lock[T]) DataLocked() (T, error) {
	var data T
	if !l.IsLocked() {
		return data, fmt.Errorf("cannot read lock file '%s' as it does not exist", l.path)
	}
	if err := fmtx.UnmarshalFile(l.path, &data); err != nil {
		return data, fmt.Errorf("cannot read lock file '%s': %w", l.path, err)
	}
	return data, nil
}

func (l Lock[T]) IsUpToDate() (bool, error) {
	if !l.IsLocked() {
		return false, nil
	}
	dataLocked, err := l.DataLocked()
	if err != nil {
		return false, err
	}
	upToDate := cmp.Equal(l.DataCurrent(), dataLocked)
	return upToDate, nil
}
