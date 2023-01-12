package base

type Opts struct {
	TmpDir           string
	ChecksumExcluded []string
}

func NewOpts() *Opts {
	return &Opts{
		TmpDir: "aem/home/tmp",
		ChecksumExcluded: []string{
			// meta files
			"**/.*/**",
			"**/.*",

			// build files
			"**/target/**",
			"**/target",
			"**/build/**",
			"**/build",
			"**/dist/**",
			"**/dist",
			"**/generated",
			"**/generated/**",
			"**/package-lock.json",

			// temporary files
			"**/node_modules/**",
			"**/node_modules",
			"**/node/**",
			"**/node",
			"**/*.log",
			"**/*.tmp",
		},
	}
}
