package replication

const (
	ReplicateJsonPath = "/bin/replicate.json"
)

type ActivateTreeOpts struct {
	StartPath         string
	DryRun            bool
	OnlyModified      bool
	OnlyActivated     bool
	IgnoreDeactivated bool
}
