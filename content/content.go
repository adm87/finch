package content

import "embed"

//go:embed embedded
var EmbeddedFiles embed.FS

const (
	DefaultImage10x10 = "embedded/default_10x10.png"
)
