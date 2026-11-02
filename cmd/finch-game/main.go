package main

import (
	"log"

	"github.com/adm87/finch/cmd/finch-game/internal"
)

func main() {
	if err := internal.Boot(); err != nil {
		log.Fatal(err)
	}
}
