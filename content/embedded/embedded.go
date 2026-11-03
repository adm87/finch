package embedded

import "embed"

const (
	EngineConfigPath = "configs/engine.yaml"
	GameConfigPath   = "configs/game.yaml"
)

//go:embed configs/*
var Configs embed.FS
