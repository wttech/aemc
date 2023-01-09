package osx

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/common/pathx"
)

type Lock[T comparable] struct {
	path        string
	dataCurrent T
	dataLocked  T
}

func NewLock[T comparable](path string, data T) Lock[T] {
	return Lock[T]{path: path, dataCurrent: data}
}

func (l Lock[T]) Data() T {
	return l.dataCurrent
}

func (l Lock[T]) Lock() error {
	err := fmtx.MarshalToFile(l.path, l.dataCurrent)
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

func (l Lock[T]) IsUpToDate() (bool, error) {
	if !l.IsLocked() {
		return false, nil
	}
	if err := fmtx.UnmarshalFile(l.path, &l.dataLocked); err != nil {
		return false, fmt.Errorf("cannot read lock file '%s': %w", l.path, err)
	}
	upToDate := cmp.Equal(l.dataCurrent, l.dataLocked)
	return upToDate, nil
}
