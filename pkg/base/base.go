package base

type Opts struct {
	TmpDir string
}

func NewOpts() *Opts {
	return &Opts{
		TmpDir: "aem/home/tmp",
	}
}
