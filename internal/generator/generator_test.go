package generator

import (
	"math/rand"
	"testing"

	"pug-generator/internal/domain"
)

// flex is a player with no preferences (can play any role).
func flex(name string) domain.Player { return domain.Player{Name: name} }

// locked is a player who only plays a single role.
func locked(name string, role domain.Role) domain.Player {
	return domain.Player{
		Name:            name,
		PreferredHeroes: map[domain.Role][]string{role: {"anything"}},
	}
}

func flexRoster(n int) []domain.Player {
	ps := make([]domain.Player, n)
	for i := range ps {
		ps[i] = flex(string(rune('A' + i)))
	}
	return ps
}

var comp122 = map[domain.Role]int{domain.RoleTank: 1, domain.RoleDPS: 2, domain.RoleSupport: 2}

func newRNG() *rand.Rand { return rand.New(rand.NewSource(1)) }

func TestRoleQueueFillsComposition(t *testing.T) {
	res, err := Generate(newRNG(), flexRoster(10), nil, Options{
		TeamCount: 2, RoleQueue: true, Composition: comp122,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(res.Teams) != 2 {
		t.Fatalf("want 2 teams, got %d", len(res.Teams))
	}
	if len(res.Bench) != 0 {
		t.Errorf("want empty bench, got %d", len(res.Bench))
	}
	for i, team := range res.Teams {
		counts := map[domain.Role]int{}
		for _, a := range team.Players {
			counts[a.Role]++
		}
		for role, want := range comp122 {
			if counts[role] != want {
				t.Errorf("team %d: role %s = %d, want %d", i, role, counts[role], want)
			}
		}
	}
}

func TestRoleQueueRespectsEligibility(t *testing.T) {
	var players []domain.Player
	players = append(players, locked("t1", domain.RoleTank), locked("t2", domain.RoleTank))
	players = append(players, locked("d1", domain.RoleDPS), locked("d2", domain.RoleDPS),
		locked("d3", domain.RoleDPS), locked("d4", domain.RoleDPS))
	players = append(players, locked("s1", domain.RoleSupport), locked("s2", domain.RoleSupport),
		locked("s3", domain.RoleSupport), locked("s4", domain.RoleSupport))

	res, err := Generate(newRNG(), players, nil, Options{
		TeamCount: 2, RoleQueue: true, Composition: comp122,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, team := range res.Teams {
		for _, a := range team.Players {
			want := eligibleRoles(a.Player)
			if !want[a.Role] {
				t.Errorf("player %s assigned role %s it cannot play", a.Player.Name, a.Role)
			}
		}
	}
}

func TestRoleQueueInfeasible(t *testing.T) {
	// Ten DPS-only players cannot fill any tank/support slot.
	var players []domain.Player
	for i := 0; i < 10; i++ {
		players = append(players, locked(string(rune('a'+i)), domain.RoleDPS))
	}
	_, err := Generate(newRNG(), players, nil, Options{
		TeamCount: 2, RoleQueue: true, Composition: comp122,
	})
	if err == nil {
		t.Fatal("expected an error for an infeasible composition")
	}
}

func TestOpenQueueBenchesExtras(t *testing.T) {
	res, err := Generate(newRNG(), flexRoster(12), nil, Options{
		TeamCount: 2, TeamSize: 5,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for i, team := range res.Teams {
		if len(team.Players) != 5 {
			t.Errorf("team %d has %d players, want 5", i, len(team.Players))
		}
		for _, a := range team.Players {
			if a.Role != "" {
				t.Errorf("open queue should not assign roles, got %q", a.Role)
			}
		}
	}
	if len(res.Bench) != 2 {
		t.Errorf("want 2 benched, got %d", len(res.Bench))
	}
}

func TestNotEnoughPlayers(t *testing.T) {
	_, err := Generate(newRNG(), flexRoster(4), nil, Options{
		TeamCount: 2, RoleQueue: true, Composition: comp122,
	})
	if err == nil {
		t.Fatal("expected an error when there are too few players")
	}
}

func TestMapIsPicked(t *testing.T) {
	maps := []domain.Map{
		{Name: "Ilios", Mode: domain.ModeControl},
		{Name: "Havana", Mode: domain.ModeEscort},
	}
	res, err := Generate(newRNG(), flexRoster(10), maps, Options{
		TeamCount: 2, RoleQueue: true, Composition: comp122,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if res.Map == nil {
		t.Fatal("expected a map to be picked")
	}
	if res.Map.Name != "Ilios" && res.Map.Name != "Havana" {
		t.Errorf("picked unexpected map %q", res.Map.Name)
	}
}

func TestNoMapWhenPoolEmpty(t *testing.T) {
	res, err := Generate(newRNG(), flexRoster(10), nil, Options{
		TeamCount: 2, RoleQueue: true, Composition: comp122,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if res.Map != nil {
		t.Errorf("expected no map, got %v", res.Map)
	}
}
