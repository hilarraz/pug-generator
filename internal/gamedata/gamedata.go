// Package gamedata exposes the static Overwatch hero roster and map list that
// are compiled into the binary.
//
// heroes.json and maps.json are generated from the OverFast API by the gendata
// command. Refresh them (e.g. after a patch adds heroes/maps) with:
//
//	go generate ./internal/gamedata
//
// See CLAUDE.md, "Keeping game data current".
package gamedata

//go:generate go run pug-generator/cmd/gendata

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"pug-generator/internal/domain"
)

//go:embed heroes.json
var heroesJSON []byte

//go:embed maps.json
var mapsJSON []byte

// Heroes returns the full embedded hero roster.
func Heroes() ([]domain.Hero, error) {
	var heroes []domain.Hero
	if err := json.Unmarshal(heroesJSON, &heroes); err != nil {
		return nil, fmt.Errorf("gamedata: decoding heroes: %w", err)
	}
	return heroes, nil
}

// Maps returns the full embedded map list.
func Maps() ([]domain.Map, error) {
	var maps []domain.Map
	if err := json.Unmarshal(mapsJSON, &maps); err != nil {
		return nil, fmt.Errorf("gamedata: decoding maps: %w", err)
	}
	return maps, nil
}
