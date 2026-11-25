//go:build windows

package osx

import (
	"fmt"

	"golang.org/x/sys/windows"
)

func DiskUsageOf(path string) (DiskUsage, error) {
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return DiskUsage{}, fmt.Errorf("cannot read disk usage of path '%s': %w", path, err)
	}
	var free, total, totalFree uint64
	if err := windows.GetDiskFreeSpaceEx(pathPtr, &free, &total, &totalFree); err != nil {
		return DiskUsage{}, fmt.Errorf("cannot read disk usage of path '%s': %w", path, err)
	}
	return DiskUsage{Free: free, Total: total}, nil
}
