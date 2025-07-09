package lomc

type ChunkType string

const (
	ChunkTypeKey   ChunkType = "key"
	ChunkTypeDelta ChunkType = "delta"
)
