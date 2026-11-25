//go:build !windows

package osx

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func DiskUsageOf(path string) (DiskUsage, error) {
	var stat unix.Statfs_t
	if err := unix.Statfs(path, &stat); err != nil {
		return DiskUsage{}, fmt.Errorf("cannot read disk usage of path '%s': %w", path, err)
	}
	blockSize := uint64(stat.Bsize) // int64 on Linux, uint32 on macOS
	return DiskUsage{
		Free:  stat.Bavail * blockSize,
		Total: stat.Blocks * blockSize,
	}, nil
}
