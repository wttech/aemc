package osx

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/wttech/aemc/pkg/common/fmtx"
	"github.com/wttech/aemc/pkg/common/pathx"
)

type Lock[T comparable] struct {
	path         string
	dataProvider func() (T, error)
}

type LockState[T comparable] struct {
	Current  T
	Locked   T
	UpToDate bool
}

func NewLock[T comparable](path string, dataProvider func() (T, error)) Lock[T] {
	return Lock[T]{path, dataProvider}
}

func (l Lock[T]) Lock() error {
	data, err := l.dataProvider()
	if err != nil {
		return fmt.Errorf("cannot compute data for lock file '%s': %w", l.path, err)
	}
	if err := fmtx.MarshalToFile(l.path, data); err != nil {
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

func (l Lock[T]) Current() (T, error) {
	return l.dataProvider()
}

func (l Lock[T]) Locked() (T, error) {
	var data T
	if !l.IsLocked() {
		return data, fmt.Errorf("cannot read lock file '%s' as it does not exist", l.path)
	}
	if err := fmtx.UnmarshalFile(l.path, &data); err != nil {
		return data, fmt.Errorf("cannot read lock file '%s': %w", l.path, err)
	}
	return data, nil
}

func (l Lock[T]) State() (LockState[T], error) {
	var zero LockState[T]
	if !l.IsLocked() {
		return zero, nil
	}
	locked, err := l.Locked()
	if err != nil {
		return zero, err
	}
	current, err := l.Current()
	if err != nil {
		return zero, err
	}
	return LockState[T]{
		Current:  current,
		Locked:   locked,
		UpToDate: cmp.Equal(current, locked),
	}, nil
}
