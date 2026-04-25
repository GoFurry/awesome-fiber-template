package core

type Request struct {
	ProjectName  string
	ModulePath   string
	Preset       string
	Capabilities []string
	Options      map[string]string
}
