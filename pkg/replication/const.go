package replication

const (
	ReplicateJsonPath = "/bin/replicate.json"
	ActivateTreePath  = "/libs/replication/treeactivation.html"
)

type ActivateTreeOpts struct {
	StartPath         string
	DryRun            bool
	OnlyModified      bool
	OnlyActivated     bool
	IgnoreDeactivated bool
}
