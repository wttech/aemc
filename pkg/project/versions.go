package project

// Versions of the tools pinned in the project files of both runners ('mise.toml', 'aemw', 'taskw').
// The release bumps AemcVersion here and in 'runner/mise/*/mise.toml', then runs 'go generate ./pkg/project'.
const (
	AemcVersion = "3.0.0"
	TaskVersion = "3.54.0"
)
