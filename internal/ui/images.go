package ui

import (
	"fyne.io/fyne/v2"

	"pug-generator/internal/gamedata"
)

// heroResource returns the embedded portrait for a hero as a Fyne resource, or
// nil when none is embedded (the card then shows a placeholder).
func heroResource(name string) fyne.Resource {
	return toResource(gamedata.HeroImage(name))
}

// mapResource returns the embedded screenshot for a map, or nil when none is
// embedded.
func mapResource(name string) fyne.Resource {
	return toResource(gamedata.MapImage(name))
}

func toResource(filename string, data []byte) fyne.Resource {
	if data == nil {
		return nil
	}
	return fyne.NewStaticResource(filename, data)
}
