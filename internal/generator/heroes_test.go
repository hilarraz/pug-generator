package generator

import (
	"math/rand"
	"testing"

	"pug-generator/internal/domain"
)

func newRNGSeed(seed int64) *rand.Rand { return rand.New(rand.NewSource(seed)) }

var tankPool = []domain.Hero{
	{Name: "Reinhardt", Role: domain.RoleTank},
	{Name: "Zarya", Role: domain.RoleTank},
	{Name: "Sigma", Role: domain.RoleTank},
}

func TestAssignHeroesUniqueWithinTeam(t *testing.T) {
	// Two teams of two tanks each; every player is happy on any of the 3 tanks.
	players := make([]domain.Player, 4)
	for i := range players {
		players[i] = domain.Player{
			Name:            string(rune('A' + i)),
			PreferredHeroes: map[domain.Role][]string{domain.RoleTank: {"Reinhardt", "Zarya", "Sigma"}},
		}
	}
	res, err := Generate(newRNG(), players, nil, Options{
		TeamCount:    2,
		RoleQueue:    true,
		Composition:  map[domain.Role]int{domain.RoleTank: 2},
		AssignHeroes: true,
		Heroes:       tankPool,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	enabled := map[string]bool{"Reinhardt": true, "Zarya": true, "Sigma": true}
	for i, team := range res.Teams {
		seen := map[string]bool{}
		for _, a := range team.Players {
			if a.Hero == "" {
				t.Errorf("team %d: player %s got no hero", i, a.Player.Name)
			}
			if !enabled[a.Hero] {
				t.Errorf("team %d: hero %q is not in the enabled pool", i, a.Hero)
			}
			if seen[a.Hero] {
				t.Errorf("team %d: hero %q assigned twice", i, a.Hero)
			}
			seen[a.Hero] = true
		}
	}
}

func TestAssignHeroesAvoidsDisliked(t *testing.T) {
	// The player flexes any tank but dislikes Reinhardt; with Zarya/Sigma free it
	// should never get Reinhardt.
	players := []domain.Player{
		{Name: "A", DislikedHeroes: []string{"Reinhardt"}},
		{Name: "B"},
	}
	for seed := int64(0); seed < 50; seed++ {
		res, err := Generate(newRNGSeed(seed), players, nil, Options{
			TeamCount:    2,
			RoleQueue:    true,
			Composition:  map[domain.Role]int{domain.RoleTank: 1},
			AssignHeroes: true,
			Heroes:       tankPool,
		})
		if err != nil {
			t.Fatalf("Generate: %v", err)
		}
		for _, team := range res.Teams {
			for _, a := range team.Players {
				if a.Player.Name == "A" && a.Hero == "Reinhardt" {
					t.Fatalf("seed %d: disliked hero Reinhardt was assigned", seed)
				}
			}
		}
	}
}

func TestNoHeroesWhenDisabled(t *testing.T) {
	res, err := Generate(newRNG(), flexRoster(4), nil, Options{
		TeamCount:   2,
		RoleQueue:   true,
		Composition: map[domain.Role]int{domain.RoleTank: 1},
		Heroes:      tankPool, // provided but AssignHeroes is false
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, team := range res.Teams {
		for _, a := range team.Players {
			if a.Hero != "" {
				t.Errorf("expected no hero when AssignHeroes is false, got %q", a.Hero)
			}
		}
	}
}

func TestAssignHeroesOpenQueue(t *testing.T) {
	players := flexRoster(4)
	res, err := Generate(newRNG(), players, nil, Options{
		TeamCount:    2,
		TeamSize:     2,
		AssignHeroes: true,
		Heroes:       tankPool,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, team := range res.Teams {
		for _, a := range team.Players {
			if a.Role != "" {
				t.Errorf("open queue should not set a role, got %q", a.Role)
			}
			if a.Hero == "" {
				t.Errorf("open queue with AssignHeroes should give a hero")
			}
		}
	}
}
