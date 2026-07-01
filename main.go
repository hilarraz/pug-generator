package main

import (
	"log"

	"pug-generator/internal/ui"
)

func main() {
	app, err := ui.New()
	if err != nil {
		log.Fatalf("pug-generator: %v", err)
	}
	app.Run()
}
