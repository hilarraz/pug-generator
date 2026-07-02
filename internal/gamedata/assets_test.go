package gamedata

import "testing"

// TestEveryHeroAndMapHasImage guards the writer/reader agreement: every hero and
// map in the embedded data must resolve, through slug.Of, to embedded image
// bytes. It catches a slug-rule change or a stale regeneration that would leave
// the UI with placeholder-only cards.
func TestEveryHeroAndMapHasImage(t *testing.T) {
	heroes, err := Heroes()
	if err != nil {
		t.Fatal(err)
	}
	for _, h := range heroes {
		if _, data := HeroImage(h.Name); data == nil {
			t.Errorf("hero %q has no embedded portrait", h.Name)
		}
	}

	maps, err := Maps()
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range maps {
		if _, data := MapImage(m.Name); data == nil {
			t.Errorf("map %q has no embedded screenshot", m.Name)
		}
	}

	if _, data := HeroImage("Definitely Not A Hero"); data != nil {
		t.Error("HeroImage(unknown): expected nil")
	}
}
