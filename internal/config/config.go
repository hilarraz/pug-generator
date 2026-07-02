// Package config holds the user-editable session state: which maps and heroes
// are in the active pool, and the roster of players. It is the document that the
// GUI saves to and loads from a JSON file.
//
// The master lists of every hero and map live in package gamedata, not here.
package config

import (
	"encoding/json"
	"fmt"
	"io"

	"pug-generator/internal/domain"
)

// currentVersion is the schema version written to new configs. Bump it (and
// migrate on load) when the on-disk shape changes incompatibly.
const currentVersion = 1

// Settings holds session-wide generation options.
type Settings struct {
	// AssignHeroes makes team generation give each player a concrete hero drawn
	// from their preferences (avoiding duplicates within a team) instead of only
	// a role. Defaults to false (role only).
	AssignHeroes bool `json:"assign_heroes"`
}

// Config is the save/loadable session state.
type Config struct {
	Version int `json:"version"`
	// EnabledMaps maps a map name to whether it is in the active pool. A name
	// that is absent defaults to enabled (see IsMapEnabled), so maps added to
	// the game data later are included by default rather than silently hidden.
	EnabledMaps map[string]bool `json:"enabled_maps"`
	// EnabledHeroes maps a hero name to whether it is in the active pool. Absent
	// names default to enabled (see IsHeroEnabled).
	EnabledHeroes map[string]bool `json:"enabled_heroes"`
	// Players is the roster registered for PUGs.
	Players []domain.Player `json:"players"`
	// Settings holds session-wide generation options.
	Settings Settings `json:"settings"`
}

// New returns an empty config with everything enabled by default.
func New() *Config {
	return &Config{
		Version:       currentVersion,
		EnabledMaps:   map[string]bool{},
		EnabledHeroes: map[string]bool{},
		Players:       []domain.Player{},
	}
}

// IsMapEnabled reports whether the named map is in the active pool. Names not
// present in the config default to enabled.
func (c *Config) IsMapEnabled(name string) bool {
	if v, ok := c.EnabledMaps[name]; ok {
		return v
	}
	return true
}

// IsHeroEnabled reports whether the named hero is in the active pool. Names not
// present in the config default to enabled.
func (c *Config) IsHeroEnabled(name string) bool {
	if v, ok := c.EnabledHeroes[name]; ok {
		return v
	}
	return true
}

// Decode reads a config from JSON. The result is normalized so its maps and
// slices are safe to mutate.
func Decode(r io.Reader) (*Config, error) {
	cfg := New()
	if err := json.NewDecoder(r).Decode(cfg); err != nil {
		return nil, fmt.Errorf("config: decode: %w", err)
	}
	cfg.normalize()
	return cfg, nil
}

// Encode writes the config as indented JSON.
func (c *Config) Encode(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(c); err != nil {
		return fmt.Errorf("config: encode: %w", err)
	}
	return nil
}

// normalize replaces nil maps/slices (e.g. from a JSON null or an older file)
// with empty, mutable ones and fills in a missing version.
func (c *Config) normalize() {
	if c.EnabledMaps == nil {
		c.EnabledMaps = map[string]bool{}
	}
	if c.EnabledHeroes == nil {
		c.EnabledHeroes = map[string]bool{}
	}
	if c.Players == nil {
		c.Players = []domain.Player{}
	}
	if c.Version == 0 {
		c.Version = currentVersion
	}
}
