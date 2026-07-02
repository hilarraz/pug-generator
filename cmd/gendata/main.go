// Command gendata regenerates the embedded game data (heroes.json, maps.json,
// and the hero/map images under assets/) from the OverFast API.
//
// It is meant to be run via `go generate ./internal/gamedata`, which sets the
// working directory to internal/gamedata so the files land in the right place.
// Non-standard maps (deathmatch, CTF, 2CP, workshop, ...) are filtered out, and
// each kept hero/map has its portrait/screenshot downloaded next to the JSON,
// named by slug.Of(name) so the UI can look it up. Downloads are best-effort: a
// missing URL or a failed fetch warns and is skipped rather than aborting.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"pug-generator/internal/domain"
	"pug-generator/internal/slug"
)

const baseURL = "https://overfast-api.tekrop.fr"

// Where the downloaded images land (relative to internal/gamedata, gendata's
// working directory when run via go generate).
const (
	assetsHeroesDir = "assets/heroes"
	assetsMapsDir   = "assets/maps"
)

// ofHero and ofMap hold only the OverFast fields we consume.
type ofHero struct {
	Name     string `json:"name"`
	Role     string `json:"role"`
	Portrait string `json:"portrait"`
}

type ofMap struct {
	Name       string   `json:"name"`
	Gamemodes  []string `json:"gamemodes"`
	Screenshot string   `json:"screenshot"`
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
	heroes, heroImgs, err := fetchHeroes()
	if err != nil {
		return err
	}
	maps, mapImgs, err := fetchMaps()
	if err != nil {
		return err
	}
	if err := writeJSON("heroes.json", heroes); err != nil {
		return err
	}
	if err := writeJSON("maps.json", maps); err != nil {
		return err
	}
	fmt.Printf("gendata: wrote %d heroes (%d portraits), %d maps (%d screenshots)\n",
		len(heroes), heroImgs, len(maps), mapImgs)
	return nil
}

func fetchHeroes() ([]domain.Hero, int, error) {
	var raw []ofHero
	if err := getJSON("/heroes", &raw); err != nil {
		return nil, 0, err
	}
	if err := resetDir(assetsHeroesDir); err != nil {
		return nil, 0, err
	}
	heroes := make([]domain.Hero, 0, len(raw))
	seen := make(map[string]string)
	imgs := 0
	for _, h := range raw {
		role, ok := roleFromOverFast[h.Role]
		if !ok {
			return nil, 0, fmt.Errorf("unknown role %q for hero %q", h.Role, h.Name)
		}
		heroes = append(heroes, domain.Hero{Name: h.Name, Role: role})
		if downloadAsset(assetsHeroesDir, h.Name, h.Portrait, seen) {
			imgs++
		}
	}
	sort.Slice(heroes, func(i, j int) bool {
		if heroes[i].Role != heroes[j].Role {
			return roleOrder(heroes[i].Role) < roleOrder(heroes[j].Role)
		}
		return heroes[i].Name < heroes[j].Name
	})
	return heroes, imgs, nil
}

func fetchMaps() ([]domain.Map, int, error) {
	var raw []ofMap
	if err := getJSON("/maps", &raw); err != nil {
		return nil, 0, err
	}
	if err := resetDir(assetsMapsDir); err != nil {
		return nil, 0, err
	}
	maps := make([]domain.Map, 0, len(raw))
	seen := make(map[string]string)
	imgs := 0
	for _, m := range raw {
		mode, ok := standardMode(m.Gamemodes)
		if !ok {
			continue // skip non-standard maps (deathmatch, CTF, 2CP, workshop, ...)
		}
		maps = append(maps, domain.Map{Name: m.Name, Mode: mode})
		if downloadAsset(assetsMapsDir, m.Name, m.Screenshot, seen) {
			imgs++
		}
	}
	sort.Slice(maps, func(i, j int) bool {
		if maps[i].Mode != maps[j].Mode {
			return modeOrder(maps[i].Mode) < modeOrder(maps[j].Mode)
		}
		return maps[i].Name < maps[j].Name
	})
	return maps, imgs, nil
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

// downloadAsset fetches one image into dir/<slug(name)><ext> and reports whether
// a file was written. It warns (but does not fail) on a duplicate slug, a
// missing URL, or a fetch error so that one bad entry doesn't abort the run.
func downloadAsset(dir, name, url string, seen map[string]string) bool {
	s := slug.Of(name)
	if prev, dup := seen[s]; dup {
		warn("slug %q collides for %q and %q; keeping the first", s, prev, name)
		return false
	}
	seen[s] = name
	if url == "" {
		warn("no image URL for %q", name)
		return false
	}
	data, err := getBytes(url)
	if err != nil {
		warn("image for %q: %v", name, err)
		return false
	}
	ext := filepath.Ext(url)
	if ext == "" {
		ext = ".png"
	}
	dst := filepath.Join(dir, s+ext)
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		warn("writing %s: %v", dst, err)
		return false
	}
	return true
}

// resetDir removes and recreates dir so stale images (e.g. for a hero that has
// left the roster, or a rename) don't linger in the embedded assets.
func resetDir(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("clearing %s: %w", dir, err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", dir, err)
	}
	return nil
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

func getBytes(url string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: unexpected status %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
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

func warn(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "gendata: warning: "+format+"\n", args...)
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
