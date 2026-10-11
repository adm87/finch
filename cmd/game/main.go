package main

import (
	"log"

	"github.com/adm87/finch/cmd/game/internal"
)

func main() {
	if err := internal.Run(); err != nil {
		log.Fatalf("error: %s", err.Error())
	}
}
