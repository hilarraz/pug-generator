// Command gendata regenerates the embedded game data
// (internal/gamedata/heroes.json and maps.json) from the OverFast API.
//
// It is meant to be run via `go generate ./internal/gamedata`, which sets the
// working directory to internal/gamedata so the files land in the right place.
// Non-standard maps (deathmatch, CTF, 2CP, workshop, ...) are filtered out.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"time"

	"pug-generator/internal/domain"
)

const baseURL = "https://overfast-api.tekrop.fr"

// ofHero and ofMap hold only the OverFast fields we consume.
type ofHero struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

type ofMap struct {
	Name      string   `json:"name"`
	Gamemodes []string `json:"gamemodes"`
}

var roleFromOverFast = map[string]domain.Role{
	"tank":    domain.RoleTank,
	"damage":  domain.RoleDPS,
	"support": domain.RoleSupport,
}

// standardModes maps OverFast gamemode keys to our GameMode, in the order we
// prefer when a map advertises several (currently every map has exactly one).
var standardModes = []struct {
	key  string
	mode domain.GameMode
}{
	{"control", domain.ModeControl},
	{"escort", domain.ModeEscort},
	{"hybrid", domain.ModeHybrid},
	{"push", domain.ModePush},
	{"flashpoint", domain.ModeFlashpoint},
	{"clash", domain.ModeClash},
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gendata:", err)
		os.Exit(1)
	}
}

func run() error {
	heroes, err := fetchHeroes()
	if err != nil {
		return err
	}
	maps, err := fetchMaps()
	if err != nil {
		return err
	}
	if err := writeJSON("heroes.json", heroes); err != nil {
		return err
	}
	if err := writeJSON("maps.json", maps); err != nil {
		return err
	}
	fmt.Printf("gendata: wrote %d heroes, %d maps\n", len(heroes), len(maps))
	return nil
}

func fetchHeroes() ([]domain.Hero, error) {
	var raw []ofHero
	if err := getJSON("/heroes", &raw); err != nil {
		return nil, err
	}
	heroes := make([]domain.Hero, 0, len(raw))
	for _, h := range raw {
		role, ok := roleFromOverFast[h.Role]
		if !ok {
			return nil, fmt.Errorf("unknown role %q for hero %q", h.Role, h.Name)
		}
		heroes = append(heroes, domain.Hero{Name: h.Name, Role: role})
	}
	sort.Slice(heroes, func(i, j int) bool {
		if heroes[i].Role != heroes[j].Role {
			return roleOrder(heroes[i].Role) < roleOrder(heroes[j].Role)
		}
		return heroes[i].Name < heroes[j].Name
	})
	return heroes, nil
}

func fetchMaps() ([]domain.Map, error) {
	var raw []ofMap
	if err := getJSON("/maps", &raw); err != nil {
		return nil, err
	}
	maps := make([]domain.Map, 0, len(raw))
	for _, m := range raw {
		mode, ok := standardMode(m.Gamemodes)
		if !ok {
			continue // skip non-standard maps (deathmatch, CTF, 2CP, workshop, ...)
		}
		maps = append(maps, domain.Map{Name: m.Name, Mode: mode})
	}
	sort.Slice(maps, func(i, j int) bool {
		if maps[i].Mode != maps[j].Mode {
			return modeOrder(maps[i].Mode) < modeOrder(maps[j].Mode)
		}
		return maps[i].Name < maps[j].Name
	})
	return maps, nil
}

func standardMode(gamemodes []string) (domain.GameMode, bool) {
	for _, sm := range standardModes {
		for _, g := range gamemodes {
			if g == sm.key {
				return sm.mode, true
			}
		}
	}
	return "", false
}

func getJSON(path string, out any) error {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(http.MethodGet, baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("GET %s: %w", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: unexpected status %s", path, resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decoding %s: %w", path, err)
	}
	return nil
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}

func roleOrder(r domain.Role) int {
	for i, x := range domain.Roles {
		if x == r {
			return i
		}
	}
	return len(domain.Roles)
}

func modeOrder(m domain.GameMode) int {
	for i, x := range domain.GameModes {
		if x == m {
			return i
		}
	}
	return len(domain.GameModes)
}
