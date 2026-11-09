package embedded

import "embed"

const GameConfigPath = "configs/game.yaml"

//go:embed configs/*
var Configs embed.FS
