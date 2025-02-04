package catalog

type Track struct {
	Path    []string
	Params  map[string]interface{}
	Depends []string
}
