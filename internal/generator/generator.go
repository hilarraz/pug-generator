// Package generator draws PUG teams from a roster of players.
//
// It is pure and deterministic given its *rand.Rand, so it can be unit-tested
// without a GUI. In role queue it assigns each player a role they can actually
// play (based on their preferred heroes) using bipartite matching, guaranteeing
// a valid composition whenever one exists.
package generator

import (
	"errors"
	"fmt"
	"math/rand"

	"pug-generator/internal/domain"
)

// Options controls how teams are generated.
type Options struct {
	// TeamCount is how many teams to build (at least 2).
	TeamCount int
	// RoleQueue toggles role-constrained generation.
	RoleQueue bool
	// Composition is the number of players per role, per team (role queue only).
	Composition map[domain.Role]int
	// TeamSize is the number of players per team (open queue only).
	TeamSize int
	// AssignHeroes gives each player a concrete hero (Assignment.Hero) drawn from
	// their preferences, unique within a team, instead of only a role.
	AssignHeroes bool
	// Heroes is the enabled hero pool to draw from when AssignHeroes is set.
	Heroes []domain.Hero
}

// Assignment is a player placed on a team, with the role they were given (empty
// in open queue) and, when AssignHeroes is set, the hero they were given.
type Assignment struct {
	Player domain.Player
	Role   domain.Role
	Hero   string
}

// Team is one generated team.
type Team struct {
	Players []Assignment
}

// Result is a generated PUG: the picked map (nil if the map pool was empty),
// the teams, and any players left on the bench.
type Result struct {
	Map   *domain.Map
	Teams []Team
	Bench []domain.Player
}

// Generate builds teams from players, picking a map at random from maps.
func Generate(rng *rand.Rand, players []domain.Player, maps []domain.Map, opts Options) (*Result, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	teamSize := opts.teamSize()
	needed := opts.TeamCount * teamSize
	if len(players) < needed {
		return nil, fmt.Errorf("generator: %d joueurs requis (%d équipes × %d), seulement %d disponibles",
			needed, opts.TeamCount, teamSize, len(players))
	}

	res := &Result{Teams: make([]Team, opts.TeamCount)}
	if len(maps) > 0 {
		m := maps[rng.Intn(len(maps))]
		res.Map = &m
	}

	if opts.RoleQueue {
		if err := assignRoleQueue(rng, players, opts, res); err != nil {
			return nil, err
		}
	} else {
		assignOpenQueue(rng, players, opts, res)
	}
	if opts.AssignHeroes {
		assignHeroes(rng, opts, res)
	}
	return res, nil
}

// assignHeroes gives every seated player a concrete hero. Within a team heroes
// are unique (as in a real match). For each player it prefers, in order: one of
// their preferred heroes for the seat, then any enabled hero for the seat that
// they don't dislike, then any enabled hero for the seat (disliked allowed as a
// last resort). A seat's role scopes the pool in role queue; in open queue the
// whole enabled pool is used. A player is left without a hero only if the pool
// for the seat is exhausted.
func assignHeroes(rng *rand.Rand, opts Options, res *Result) {
	byRole := make(map[domain.Role][]string)
	all := make([]string, 0, len(opts.Heroes))
	for _, h := range opts.Heroes {
		byRole[h.Role] = append(byRole[h.Role], h.Name)
		all = append(all, h.Name)
	}
	for ti := range res.Teams {
		used := make(map[string]bool)
		for pi := range res.Teams[ti].Players {
			a := &res.Teams[ti].Players[pi]
			pool := all
			if opts.RoleQueue {
				pool = byRole[a.Role]
			}
			if hero := pickHero(rng, a.Player, a.Role, opts.RoleQueue, pool, used); hero != "" {
				a.Hero = hero
				used[hero] = true
			}
		}
	}
}

func pickHero(rng *rand.Rand, p domain.Player, role domain.Role, roleQueue bool, pool []string, used map[string]bool) string {
	inPool := make(map[string]bool, len(pool))
	for _, n := range pool {
		inPool[n] = true
	}
	disliked := make(map[string]bool)
	for _, n := range p.DislikedHeroes {
		disliked[n] = true
	}

	var preferred []string
	if roleQueue {
		preferred = p.PreferredHeroes[role]
	} else {
		for _, r := range domain.Roles {
			preferred = append(preferred, p.PreferredHeroes[r]...)
		}
	}

	// Tier 1: a preferred hero that is enabled, not disliked, not taken.
	if h := choose(rng, preferred, inPool, disliked, used); h != "" {
		return h
	}
	// Tier 2: any enabled hero for the seat, not disliked, not taken.
	if h := choose(rng, pool, nil, disliked, used); h != "" {
		return h
	}
	// Tier 3: any enabled hero for the seat, not taken (disliked allowed).
	return choose(rng, pool, nil, nil, used)
}

// choose returns a random name from candidates that is in mustBeIn (when
// non-nil) and in neither exclude set, or "" if none qualifies.
func choose(rng *rand.Rand, candidates []string, mustBeIn, excludeA, excludeB map[string]bool) string {
	var ok []string
	for _, n := range candidates {
		if mustBeIn != nil && !mustBeIn[n] {
			continue
		}
		if excludeA[n] || excludeB[n] {
			continue
		}
		ok = append(ok, n)
	}
	if len(ok) == 0 {
		return ""
	}
	return ok[rng.Intn(len(ok))]
}

func assignOpenQueue(rng *rand.Rand, players []domain.Player, opts Options, res *Result) {
	order := rng.Perm(len(players))
	needed := opts.TeamCount * opts.TeamSize
	for i := 0; i < needed; i++ {
		team := i / opts.TeamSize
		res.Teams[team].Players = append(res.Teams[team].Players, Assignment{Player: players[order[i]]})
	}
	for i := needed; i < len(players); i++ {
		res.Bench = append(res.Bench, players[order[i]])
	}
}

func assignRoleQueue(rng *rand.Rand, players []domain.Player, opts Options, res *Result) error {
	// One slot per required seat, e.g. 2 teams of 1-2-2 -> 10 slots.
	type slot struct {
		team int
		role domain.Role
	}
	var slots []slot
	for team := 0; team < opts.TeamCount; team++ {
		for _, role := range domain.Roles {
			for i := 0; i < opts.Composition[role]; i++ {
				slots = append(slots, slot{team: team, role: role})
			}
		}
	}

	// Eligible player indices per role, shuffled so matching varies each run.
	eligible := make(map[domain.Role][]int)
	for i, p := range players {
		for role := range eligibleRoles(p) {
			eligible[role] = append(eligible[role], i)
		}
	}
	for role := range eligible {
		list := eligible[role]
		rng.Shuffle(len(list), func(a, b int) { list[a], list[b] = list[b], list[a] })
	}

	// Bipartite matching (Kuhn's algorithm): give each slot a distinct player
	// eligible for that slot's role.
	slotPlayer := make([]int, len(slots))
	playerSlot := make([]int, len(players))
	for i := range slotPlayer {
		slotPlayer[i] = -1
	}
	for i := range playerSlot {
		playerSlot[i] = -1
	}

	var visited []bool
	var tryAssign func(s int) bool
	tryAssign = func(s int) bool {
		for _, p := range eligible[slots[s].role] {
			if visited[p] {
				continue
			}
			visited[p] = true
			if playerSlot[p] == -1 || tryAssign(playerSlot[p]) {
				slotPlayer[s] = p
				playerSlot[p] = s
				return true
			}
		}
		return false
	}
	for s := range slots {
		visited = make([]bool, len(players))
		if !tryAssign(s) {
			return errors.New("generator: composition impossible avec ce roster " +
				"(pas assez de joueurs éligibles pour un rôle — ajuste les préférences ou passe en open queue)")
		}
	}

	for s, sl := range slots {
		res.Teams[sl.team].Players = append(res.Teams[sl.team].Players, Assignment{
			Player: players[slotPlayer[s]],
			Role:   sl.role,
		})
	}
	for p := range players {
		if playerSlot[p] == -1 {
			res.Bench = append(res.Bench, players[p])
		}
	}
	return nil
}

// eligibleRoles returns the roles a player can fill. A player with at least one
// preferred hero is eligible for exactly the roles they have picks in; a player
// with no preferences at all is treated as flex (every role).
func eligibleRoles(p domain.Player) map[domain.Role]bool {
	res := make(map[domain.Role]bool)
	for _, r := range domain.Roles {
		if len(p.PreferredHeroes[r]) > 0 {
			res[r] = true
		}
	}
	if len(res) == 0 {
		for _, r := range domain.Roles {
			res[r] = true
		}
	}
	return res
}

func (o Options) validate() error {
	if o.TeamCount < 2 {
		return errors.New("generator: il faut au moins 2 équipes")
	}
	if o.RoleQueue {
		total := 0
		for _, r := range domain.Roles {
			c := o.Composition[r]
			if c < 0 {
				return errors.New("generator: la composition ne peut pas être négative")
			}
			total += c
		}
		if total < 1 {
			return errors.New("generator: la composition doit contenir au moins 1 joueur")
		}
		return nil
	}
	if o.TeamSize < 1 {
		return errors.New("generator: la taille d'équipe doit être d'au moins 1")
	}
	return nil
}

func (o Options) teamSize() int {
	if !o.RoleQueue {
		return o.TeamSize
	}
	total := 0
	for _, r := range domain.Roles {
		total += o.Composition[r]
	}
	return total
}
