package content

import "embed"

//go:embed embedded
var EmbeddedFiles embed.FS

const (
	GameConfig        = "embedded/conf/game_config.yaml"
	DefaultImage10x10 = "embedded/default_10x10.png"
)
