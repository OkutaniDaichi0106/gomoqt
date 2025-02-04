package catalog

type Root struct {
	Version   uint32
	Packaging Packaging
	Tracks    []Track
}

type Packaging string

const (
	CMAF Packaging = "cmaf"
	LOC  Packaging = "loc"
)
