package osx

// DiskUsage describes the space of the file system holding a path.
type DiskUsage struct {
	// Free is the space available to the current user (not to root).
	Free  uint64
	Total uint64
}
