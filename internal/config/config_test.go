package config

import (
	"bytes"
	"testing"

	"pug-generator/internal/domain"
)

func TestDefaultsEnabled(t *testing.T) {
	c := New()
	if !c.IsMapEnabled("Some Map") {
		t.Error("unknown map should default to enabled")
	}
	if !c.IsHeroEnabled("Some Hero") {
		t.Error("unknown hero should default to enabled")
	}
}

func TestRoundTrip(t *testing.T) {
	c := New()
	c.EnabledMaps["Havana"] = false
	c.EnabledHeroes["Genji"] = false
	c.Players = append(c.Players, domain.Player{
		Name:            "Alice",
		PreferredHeroes: map[domain.Role][]string{domain.RoleDPS: {"Tracer"}},
		PreferredMaps:   []string{"Ilios"},
	})

	var buf bytes.Buffer
	if err := c.Encode(&buf); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	got, err := Decode(&buf)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got.IsMapEnabled("Havana") {
		t.Error("Havana should be disabled after round-trip")
	}
	if got.IsHeroEnabled("Genji") {
		t.Error("Genji should be disabled after round-trip")
	}
	if len(got.Players) != 1 || got.Players[0].Name != "Alice" {
		t.Fatalf("player not round-tripped: %+v", got.Players)
	}
	if dps := got.Players[0].PreferredHeroes[domain.RoleDPS]; len(dps) != 1 || dps[0] != "Tracer" {
		t.Errorf("preferred heroes not round-tripped: %v", dps)
	}
}

func TestDecodeNormalizesNull(t *testing.T) {
	got, err := Decode(bytes.NewBufferString(
		`{"version":1,"enabled_maps":null,"enabled_heroes":null,"players":null}`))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	// Mutating must not panic on a nil map.
	got.EnabledMaps["X"] = true
	if got.Players == nil {
		t.Error("players should be normalized to non-nil")
	}
}
