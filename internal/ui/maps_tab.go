package ui

import (
	"sort"

	"pug-generator/internal/domain"
)

// newMapsPool builds the map-pool selector, with one category per game mode.
func (a *App) newMapsPool() *poolView {
	byMode := make(map[domain.GameMode][]string)
	for _, m := range a.maps {
		byMode[m.Mode] = append(byMode[m.Mode], m.Name)
	}

	var cats []poolCategory
	for _, mode := range domain.GameModes {
		names := byMode[mode]
		if len(names) == 0 {
			continue
		}
		sort.Strings(names)
		cats = append(cats, poolCategory{
			name:  string(mode),
			color: modeColor(mode),
			items: names,
		})
	}

	return newPoolView(cats,
		func(name string) bool { return a.cfg.IsMapEnabled(name) },
		func(name string, on bool) { a.cfg.EnabledMaps[name] = on },
	)
}
