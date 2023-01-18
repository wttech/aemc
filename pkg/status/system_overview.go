package status

type SystemOverview struct {
	Instance struct {
		AemVersion string `json:"Adobe Experience Manager"`
		RunModes   string `json:"Run Modes"`
		UpSince    string `json:"Instance Up Since"`
	} `json:"Instance"`
	Repository struct {
		OakVersion    string `json:"Apache Jackrabbit Oak"`
		NodeStore     string `json:"Node Store"`
		Size          string `json:"Repository Size"`
		FileDataStore string `json:"File Data Store"`
	} `json:"Repository"`
	MaintenanceTasks struct {
		Succeeded    string `json:"Succeeded"`
		NeverRun     string `json:"Never Run"`
		NotScheduled string `json:"Not scheduled"`
	} `json:"Maintenance Tasks"`
	HealthChecks        map[string]string `json:"Health Checks"`
	SystemInformation   map[string]string `json:"System Information"`
	EstimatedNodeCounts map[string]string `json:"Estimated Node Counts"`
}
