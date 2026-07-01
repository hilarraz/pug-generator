package gamedata

import (
	"testing"

	"pug-generator/internal/domain"
)

func TestHeroesLoad(t *testing.T) {
	heroes, err := Heroes()
	if err != nil {
		t.Fatalf("Heroes(): %v", err)
	}
	if len(heroes) == 0 {
		t.Fatal("Heroes(): empty roster")
	}

	valid := map[domain.Role]bool{
		domain.RoleTank: true, domain.RoleDPS: true, domain.RoleSupport: true,
	}
	seen := make(map[string]bool)
	for _, h := range heroes {
		if h.Name == "" {
			t.Error("hero with empty name")
		}
		if !valid[h.Role] {
			t.Errorf("hero %q has invalid role %q", h.Name, h.Role)
		}
		if seen[h.Name] {
			t.Errorf("duplicate hero %q", h.Name)
		}
		seen[h.Name] = true
	}
}

func TestMapsLoad(t *testing.T) {
	maps, err := Maps()
	if err != nil {
		t.Fatalf("Maps(): %v", err)
	}
	if len(maps) == 0 {
		t.Fatal("Maps(): empty list")
	}

	valid := make(map[domain.GameMode]bool)
	for _, m := range domain.GameModes {
		valid[m] = true
	}
	seen := make(map[string]bool)
	for _, m := range maps {
		if m.Name == "" {
			t.Error("map with empty name")
		}
		if !valid[m.Mode] {
			t.Errorf("map %q has invalid mode %q", m.Name, m.Mode)
		}
		if seen[m.Name] {
			t.Errorf("duplicate map %q", m.Name)
		}
		seen[m.Name] = true
	}
}
