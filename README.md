# Overwatch PUG Generator

[![Build Windows](https://github.com/hilarraz/pug-generator/actions/workflows/build-windows.yml/badge.svg)](https://github.com/hilarraz/pug-generator/actions/workflows/build-windows.yml)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)](go.mod)
[![Fyne](https://img.shields.io/badge/GUI-Fyne%20v2-1E90FF)](https://fyne.io)

A desktop app that randomly builds fair, preference-aware teams for **Overwatch
Pick-Up Games (PUGs)** played in custom lobbies. Configure your maps, heroes and
players once, save the setup, and draw fresh teams every game night.

> The application UI is in **French** (its target audience). This document, the
> code and the comments are in English.

---

## Table of contents

- [The problem](#the-problem)
- [The solution](#the-solution)
- [Features](#features)
- [Tech stack](#tech-stack)
- [Getting started](#getting-started)
- [Building](#building)
- [How it works](#how-it-works)
- [Project structure](#project-structure)
- [Keeping game data current](#keeping-game-data-current)
- [Roadmap](#roadmap)
- [License](#license)

---

## The problem

Organizing a Pick-Up Game night in Overwatch is fiddly. Someone has to:

- pick a map from the pool everyone agreed on,
- split a variable number of players into balanced teams,
- respect **role queue** (a valid Tank / DPS / Support composition per team),
- account for who actually enjoys playing what — nobody wants to be forced onto a
  hero or a map they dislike,
- and redo all of it, by hand, every single round.

Doing this manually is slow, error-prone, and biased. It also gets thrown away
after each session, so the group re-negotiates the same pools and preferences
every time.

## The solution

**Overwatch PUG Generator** turns that manual chore into one click.

The organizer configures three pools once — **maps**, **heroes**, and a **roster
of players** (each with per-role hero preferences and liked/disliked maps) — and
the app draws teams from that configuration. In role queue it guarantees a valid
composition whenever one exists, assigning each player a role they can actually
play, and it can optionally hand each player a concrete hero drawn from their
preferences.

The whole session is a **single JSON document** you can save and reload, so a
group keeps and tweaks its setup between game nights instead of rebuilding it.

## Features

- **Map pool** — enable/disable maps from image cards, grouped by game mode, with
  select-all / clear-all.
- **Hero pool** — same card-based selection, grouped by role.
- **Player roster** — per-player, image-based preferences with sensible caps:
  preferred heroes (3 per role), disliked heroes (3), preferred maps (3), disliked
  maps (3), all picked from a card gallery.
- **Configurable team generation** — team count, role queue on/off, and per-role
  composition, with role-aware assignment (bipartite matching on preferred heroes)
  plus a random map pick from the enabled pool.
- **Optional hero assignment** — a session setting draws each seated player a
  concrete hero from their preferences (unique per team, avoiding disliked ones).
- **Save / load** — the entire session serializes to one JSON file.
- **Offline-first** — the hero roster, map list and all artwork are embedded in
  the binary; no network needed to run.
- **Single self-contained executable** — ships as one Windows `.exe`.

## Tech stack

| Layer            | Choice                                                       |
|------------------|-------------------------------------------------------------|
| Language         | [Go](https://go.dev) 1.26                                   |
| GUI              | [Fyne](https://fyne.io) v2 — pure-Go widget toolkit         |
| Persistence      | One JSON file (the session config)                          |
| Static game data | Hero/map lists + images embedded via `go:embed`             |
| Game data source | [OverFast API](https://overfast-api.tekrop.fr) (dev-time)   |
| Dev platform     | macOS (arm64) · **Target:** Windows (amd64)                 |

> **CGO note:** Fyne renders via OpenGL, so builds require CGO and a C toolchain.
> This is transparent when building *on* macOS or *on* Windows, but matters for
> cross-compilation (see [Building](#building)).

## Getting started

### Prerequisites

- **Go 1.26+** — the version is pinned in [`mise.toml`](mise.toml); if you use
  [mise](https://mise.jdx.dev), `mise install` sets it up.
- A **C toolchain** (Fyne uses CGO):
  - macOS: Xcode Command Line Tools (`xcode-select --install`).
  - Windows: a mingw-w64 `gcc` on `PATH`.

### Run in development (macOS)

```bash
go run .
```

### Try it without building

Every push to `main` builds a Windows executable in CI. Grab it without any local
toolchain:

1. Open the **[Actions](https://github.com/hilarraz/pug-generator/actions)** tab.
2. Pick the latest **build-windows** run.
3. Download the `pug-generator-windows-amd64` artifact and run the `.exe`.

## Building

```bash
# Local binary (host OS)
go build -o bin/pug-generator .

# Format / vet / test
go fmt ./...
go vet ./...
go test ./...
```

### Cross-compile to Windows from macOS

A plain `GOOS=windows go build` will **not** work: Fyne needs CGO, so
cross-compiling requires a mingw-w64 toolchain. The path of least resistance from
a Mac is [`fyne-cross`](https://github.com/fyne-io/fyne-cross) (Docker):

```bash
go install github.com/fyne-io/fyne-cross@latest
fyne-cross windows -arch=amd64   # output in ./fyne-cross/bin/windows-amd64/
```

### CI (recommended, no local Docker or Windows needed)

[`.github/workflows/build-windows.yml`](.github/workflows/build-windows.yml)
builds the Windows `.exe` on a GitHub-hosted Windows runner (which already has the
required C toolchain) on every push to `main`/`master` and on manual dispatch. The
binary is built with `-H=windowsgui` (no console window) and published as the
`pug-generator-windows-amd64` artifact.

## How it works

The design keeps **two data layers deliberately separate**:

1. **Game data** (`internal/gamedata`) — the read-only master lists of every hero
   and map in the game, plus their images, embedded in the binary. The same for
   everyone; not user data.
2. **Session config** (`internal/config`) — the user's *choices*: which maps and
   heroes are in the active pool, and the player roster. This is what gets saved to
   and loaded from JSON.

Selection is stored as `map[string]bool` (name → enabled). A name **absent** from
the map defaults to **enabled**, so when a new hero or map is added to the embedded
game data later, existing saved configs treat it as part of the pool instead of
silently hiding it.

Team generation (`internal/generator`) is a **pure, deterministic** function given
its random source, so it is fully unit-tested without a GUI:

- **Role queue** — assigns each player a role they can actually play, using
  bipartite matching over their preferred heroes; this guarantees a valid
  composition whenever one exists.
- **Open queue** — fills teams up to a fixed size.
- **Hero assignment** (optional) — draws each player a concrete hero from their
  preferences, unique within a team and avoiding their disliked heroes.
- **Map pick** — a random map from the enabled pool.

The GUI reads the embedded game data once at startup, holds a single in-memory
`*config.Config`, mutates it as the user clicks, and serializes it on save.

## Project structure

```
main.go                 Entry point: builds the UI app and runs it.
cmd/
  gendata/              Dev tool: regenerates the gamedata JSON + images from OverFast.
internal/
  domain/               Core, dependency-free domain types (Role, Hero, Map, Player).
  slug/                 Shared name → filesystem-safe slug (writer/reader agree).
  gamedata/             Static Overwatch data embedded in the binary (JSON + assets/).
  config/               The user-editable, save/loadable session state.
  generator/            Team generation. Pure and unit-tested (no GUI).
  ui/                   Fyne GUI. All user-facing strings are in French.
```

### Domain model

- **`Role`** — `Tank`, `DPS`, `Support`.
- **`Hero`** — `{ Name, Role }`.
- **`GameMode`** — `Control`, `Escort`, `Hybrid`, `Push`, `Flashpoint`, `Clash`.
- **`Map`** — `{ Name, Mode }`.
- **`Player`** — `{ Name, PreferredHeroes (per role), DislikedHeroes, PreferredMaps,
  DislikedMaps }`.
- **`Config`** — `{ Version, EnabledMaps, EnabledHeroes, Players, Settings }`,
  serialized as the session JSON.

## Keeping game data current

`internal/gamedata/heroes.json`, `maps.json` and the images under
`internal/gamedata/assets/{heroes,maps}/` are **generated** from the community
[OverFast API](https://overfast-api.tekrop.fr) by `cmd/gendata`. Refresh them after
a patch:

```bash
go generate ./internal/gamedata
```

The tool keeps only maps whose game mode is a standard PUG mode (control, escort,
hybrid, push, flashpoint, clash), maps OverFast's `damage` role to `DPS`, and
downloads each portrait/screenshot into `assets/` named `slug.Of(name).<ext>` so
the UI can find it. The result (JSON + images) is committed so the app stays
offline-first. To swap in your own artwork, drop a `<slug>.png`/`.jpg` into
`assets/heroes` or `assets/maps` and rebuild.

> After refreshing, review the roster: OverFast may include very recent or teased
> heroes that aren't in standard play. They can be unchecked in the **Héros** tab,
> or hard-filtered in `cmd/gendata`.

## Roadmap

- [x] Foundation: pools + player roster + JSON save/load GUI.
- [x] Team generation: configurable role-queue / open-queue draw with role-aware
      assignment + random map pick. Pure and unit-tested.
- [x] Hero assignment: draw each player a concrete hero from their preferences
      *(partial — via the **Paramètres** `AssignHeroes` option)*.
- [ ] **Hero ban**: per-team hero bans before a round, removing them from that
      round's available heroes.
- [ ] **Preference-aware map pick & balancing**: weight the map pick by players'
      liked/disliked maps, and add an optional per-player skill rating to balance
      teams.

## License

Released under the [MIT License](LICENSE) — free to use, modify and distribute,
provided the copyright notice is kept.
