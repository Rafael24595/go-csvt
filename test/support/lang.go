package support

type Lang struct {
	Name       string
	Release    Release
	Tags       []string
	Attributes map[string]string
}

type Release struct {
	Version string
	Stable  bool
}
