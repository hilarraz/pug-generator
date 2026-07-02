# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Overwatch PUG Generator** is a Windows desktop application that randomly builds
teams for Pick-Up Games (PUGs) played in Overwatch custom lobbies. The organizer
configures three pools — maps, heroes, and a roster of players (each with per-role
hero preferences and liked/disliked maps) — and the app draws teams from that
configuration.

The entire session is a single JSON document the user can **save and reload**, so a
group can keep and tweak its setup between game nights.

**Status — early foundation.** What exists today:

- A Fyne GUI with five tabs: **Maps**, **Héros**, **Joueurs**, **PUG**,
  **Paramètres**.
- Map pool and hero pool selection as image cards (per-item toggle, select-all /
  clear-all).
- Player roster with capped, image-based preferences: preferred heroes (3 per
  role), disliked heroes (3), preferred maps (3), disliked maps (3) — picked from
  a card gallery.
- Configurable team generation (team count, role queue on/off, per-role
  composition) with role-aware assignment, plus a random map pick. An optional
  setting (**Paramètres**) also draws each player a concrete hero from their
  preferences (unique per team, avoiding disliked).
- Load/Save of the whole session to a JSON file.
- Embedded game data (incl. hero/map images) generated from the OverFast API via
  `go generate`.

Planned (see [Roadmap](#roadmap)): the in-match hero-ban flow (each team bans
heroes before a round), and skill-aware balancing.

## Tech Stack

- **Language:** Go 1.26 (see `mise.toml` / `go.mod`).
- **GUI:** [Fyne](https://fyne.io) v2 (v2.7.4) — pure-Go widget toolkit, ships as a
  single self-contained binary.
- **Dev platform:** macOS (arm64). **Target platform:** Windows (amd64).
- **Persistence:** one JSON file (the session config), loaded/saved from the GUI.
- **Static game data:** the hero roster and map list are compiled into the binary
  with `go:embed` (`internal/gamedata`).

> **CGO note:** Fyne renders via OpenGL, so builds require CGO and a C toolchain.
> This is transparent when building *on* macOS or *on* Windows, but matters for
> cross-compilation (see Commands).

## Commands

```bash
# Run in development (on macOS)
go run .

# Build a local binary
go build -o bin/pug-generator .

# Format / vet / test
go fmt ./...
go vet ./...
go test ./...

# Cross-compile to Windows from macOS (recommended: fyne-cross, uses Docker)
go install github.com/fyne-io/fyne-cross@latest
fyne-cross windows -arch=amd64        # output in ./fyne-cross/bin/windows-amd64/

# Refresh the embedded hero/map data from the OverFast API
go generate ./internal/gamedata
```

> A plain `GOOS=windows go build` will **not** work out of the box: Fyne needs CGO,
> so cross-compiling requires a mingw-w64 toolchain. `fyne-cross` (Docker) is the
> path of least resistance from a Mac; alternatively build directly on Windows.

**CI:** `.github/workflows/build-windows.yml` builds the Windows `.exe` on a
GitHub-hosted Windows runner (which has the required C toolchain) on every push to
`main`/`master` and on manual dispatch. Download it from the run's **Artifacts**
(`pug-generator-windows-amd64`). This needs no local Docker or Windows machine.

## Architecture

```
main.go                     Entry point: builds the UI app and runs it.
cmd/
  gendata/                  Dev tool: regenerates the gamedata JSON + images from OverFast.
internal/
  domain/                   Core, dependency-free domain types.
    role.go                 Role (Tank / DPS / Support).
    hero.go                 Hero (name + role).
    map.go                  Map (name + GameMode) and GameMode constants.
    player.go               Player (name + per-role hero prefs + map prefs).
  slug/                     Shared name → filesystem-safe slug (writer/reader agree).
    slug.go                 slug.Of, used by gendata (names files) and gamedata (looks them up).
  gamedata/                 Static Overwatch data embedded in the binary.
    gamedata.go             go:embed JSON loaders + go:generate directive.
    assets.go               go:embed assets/ + HeroImage/MapImage(name) lookup by slug.
    heroes.json             The hero roster (generated — see "Keeping game data current").
    maps.json               The map list (generated — see "Keeping game data current").
    assets/heroes/          Hero portraits, assets/maps/ map screenshots (generated, named by slug).
  config/                   The user-editable, save/loadable session state.
    config.go               Config struct + Decode/Encode (io.Reader/Writer).
  generator/                Team generation. Pure and unit-tested (no GUI).
    generator.go            Configurable draw: role queue (bipartite matching) or open queue.
  ui/                       Fyne GUI. All user-facing strings are in French.
    app.go                  App struct: window, toolbar (Load/Save), tabs, wiring.
    maps_tab.go             Map pool tab (cards grouped by game mode).
    heroes_tab.go           Hero pool tab (cards grouped by role).
    poolview.go             Shared master/detail pool selector (categories + item cards).
    poolcard.go             Tappable image+name card with included/excluded look.
    prefselect.go           Capped image multi-select (chips + card-gallery picker).
    images.go               Wraps embedded hero/map images as Fyne resources.
    players_tab.go          Player roster tab (master/detail: list + preference editor).
    pug_tab.go              PUG tab: generation controls + results display.
    settings_tab.go         Paramètres tab: session-wide generation options.
```

### Data flow & the two data layers

There are deliberately **two** separate data layers; keep them separate:

1. **Game data** (`internal/gamedata`) — the *master* lists of every hero and map in
   the game. Read-only, embedded, the same for everyone. This is not user data.
2. **Session config** (`internal/config`) — the user's *choices*: which maps/heroes
   are in the active pool, and the player roster. This is what gets saved to / loaded
   from JSON.

The config stores selection as `map[string]bool` (name → enabled). A name **absent**
from the map defaults to **enabled** (`Config.IsMapEnabled` / `IsHeroEnabled`). This
matters: when a new hero or map is added to the embedded game data later, existing
saved configs automatically treat it as part of the pool instead of silently hiding
it.

The UI (`internal/ui`) reads the embedded game data once at startup, holds a single
`*config.Config` in memory, mutates it as the user clicks, and serializes it on Save.

## Domain Model

- **`Role`** — `Tank`, `DPS`, `Support` (`domain.Roles` is the canonical ordered list).
- **`Hero`** — `{ Name string; Role Role }`.
- **`GameMode`** — `Control`, `Escort`, `Hybrid`, `Push`, `Flashpoint`, `Clash`
  (`domain.GameModes` is the canonical ordered list).
- **`Map`** — `{ Name string; Mode GameMode }`.
- **`Player`** — `{ Name; PreferredHeroes map[Role][]string; DislikedHeroes []string;
  PreferredMaps []string; DislikedMaps []string }`. The UI caps each list (3 per role
  for preferred heroes; 3 for the flat lists). Heroes have a fixed role, so disliked
  heroes are a single flat list rather than per-role.
- **`Config`** — `{ Version int; EnabledMaps, EnabledHeroes map[string]bool;
  Players []Player; Settings Settings }`, where `Settings` = `{ AssignHeroes bool }`.
  Serialized as the session JSON.

## Conventions

- Idiomatic Go: exported identifiers documented, errors wrapped with `%w` and a
  package prefix (e.g. `"config: decode: %w"`), `internal/` for non-public packages.
- `domain` has **no** dependencies (not even Fyne). Keep it that way — it's the shared
  vocabulary between the data, config, and UI layers.
- **Code and comments in English; user-facing UI strings in French.**
- Prefer small, testable pure functions for logic (config, and later team generation)
  so they can be unit-tested without spinning up a GUI.

### Keeping game data current

`internal/gamedata/heroes.json`, `maps.json`, and the images under
`internal/gamedata/assets/{heroes,maps}/` are **generated** from the community
[OverFast API](https://overfast-api.tekrop.fr) by `cmd/gendata`. Refresh them after a
patch with `go generate ./internal/gamedata`. The tool keeps only maps whose game mode
is a standard PUG mode (control, escort, hybrid, push, flashpoint, clash) and maps
OverFast's `damage` role to `DPS`. For each kept hero/map it downloads the OverFast
portrait/screenshot into `assets/`, named `slug.Of(name).<ext>` so the UI can find it;
missing/failed downloads warn and are skipped rather than aborting the run. The result
(JSON + images) is committed so the app stays offline-first — the API is only hit when
regenerating.

To swap in your own artwork, drop a `<slug>.png`/`.jpg` into `assets/heroes` or
`assets/maps` (e.g. `dva.png`); a rebuild re-embeds it. Names with no image render as a
plain card with just the label.

Review the roster after refreshing: OverFast may include very recent or teased heroes
that aren't in standard play. Unwanted heroes can also simply be unchecked in the Héros
tab; if a hard filter is needed, add it in `cmd/gendata`.

## Roadmap

1. **(done)** Foundation: pools + player roster + JSON save/load GUI.
2. **(done)** Team generation (`internal/generator`): configurable role-queue /
   open-queue draw with role-aware assignment (bipartite matching on preferred
   heroes) + a random map pick from the enabled pool. Pure, unit-tested.
3. **(partial)** Hero assignment: the **Paramètres** `AssignHeroes` option draws each
   seated player a concrete hero from their preferences (unique per team, avoiding
   disliked). See `assignHeroes` in `internal/generator`.
4. **Hero ban**: per-team hero bans before a round, removing them from that round's
   available heroes.
5. **Preference-aware map pick & balancing**: weight the map pick by players'
   preferred/disliked maps (currently uniform random), and add an optional per-player
   skill rating to balance teams (the model has no skill data today, so team
   assignment within eligibility is random).

---

## Working Guidelines

Behavioral guidelines to reduce common LLM coding mistakes.

**Tradeoff:** These guidelines bias toward caution over speed. For trivial tasks, use judgment.

### Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

### Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

### Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

### Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.

**These guidelines are working if:** fewer unnecessary changes in diffs, fewer rewrites due to overcomplication, and clarifying questions come before implementation rather than after mistakes.
