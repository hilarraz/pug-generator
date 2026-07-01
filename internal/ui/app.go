// Package ui implements the Fyne desktop GUI. All user-facing strings are in
// French; code and comments are in English (see CLAUDE.md, Conventions).
package ui

import (
	"sort"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"pug-generator/internal/config"
	"pug-generator/internal/domain"
	"pug-generator/internal/gamedata"
)

// App wires the Fyne UI to the embedded game data and the in-memory session
// config. A single *config.Config is mutated as the user interacts and is
// serialized on Save.
type App struct {
	fyneApp fyne.App
	win     fyne.Window

	heroes []domain.Hero
	maps   []domain.Map

	cfg *config.Config

	mapsView    *poolView
	heroesView  *poolView
	playersView *playersTab
	pugView     *pugTab
}

// New builds the application, loading the embedded game data.
func New() (*App, error) {
	heroes, err := gamedata.Heroes()
	if err != nil {
		return nil, err
	}
	maps, err := gamedata.Maps()
	if err != nil {
		return nil, err
	}

	a := &App{
		fyneApp: fyneapp.NewWithID("com.puggenerator.app"),
		heroes:  heroes,
		maps:    maps,
		cfg:     config.New(),
	}
	a.fyneApp.Settings().SetTheme(overwatchTheme{})
	a.win = a.fyneApp.NewWindow("Overwatch PUG Generator")
	a.win.Resize(fyne.NewSize(960, 680))
	a.win.SetContent(a.buildUI())
	return a, nil
}

// Run shows the main window and starts the event loop. It blocks until the
// window is closed.
func (a *App) Run() {
	a.win.ShowAndRun()
}

func (a *App) buildUI() fyne.CanvasObject {
	a.mapsView = a.newMapsPool()
	a.heroesView = a.newHeroesPool()
	a.playersView = a.newPlayersTab()
	a.pugView = a.newPugTab()

	tabs := container.NewAppTabs(
		container.NewTabItem("Maps", a.mapsView.root),
		container.NewTabItem("Héros", a.heroesView.root),
		container.NewTabItem("Joueurs", a.playersView.root),
		container.NewTabItem("PUG", a.pugView.content),
	)

	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.FolderOpenIcon(), a.onLoad),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), a.onSave),
	)

	return container.NewBorder(toolbar, nil, nil, nil, tabs)
}

// onLoad opens a JSON session file and replaces the current config with it.
func (a *App) onLoad() {
	dialog.ShowFileOpen(func(rc fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, a.win)
			return
		}
		if rc == nil {
			return // cancelled
		}
		defer rc.Close()

		cfg, err := config.Decode(rc)
		if err != nil {
			dialog.ShowError(err, a.win)
			return
		}
		a.cfg = cfg
		a.refreshAll()
	}, a.win)
}

// onSave writes the current config to a JSON session file.
func (a *App) onSave() {
	dialog.ShowFileSave(func(wc fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, a.win)
			return
		}
		if wc == nil {
			return // cancelled
		}
		defer wc.Close()

		if err := a.cfg.Encode(wc); err != nil {
			dialog.ShowError(err, a.win)
		}
	}, a.win)
}

// refreshAll re-syncs every tab with the current config (used after a load).
func (a *App) refreshAll() {
	a.mapsView.refresh()
	a.heroesView.refresh()
	a.playersView.refresh(a)
}

// heroNamesByRole returns the sorted hero names for a role.
func (a *App) heroNamesByRole(role domain.Role) []string {
	var names []string
	for _, h := range a.heroes {
		if h.Role == role {
			names = append(names, h.Name)
		}
	}
	sort.Strings(names)
	return names
}

// mapNames returns every map name, sorted.
func (a *App) mapNames() []string {
	names := make([]string, 0, len(a.maps))
	for _, m := range a.maps {
		names = append(names, m.Name)
	}
	sort.Strings(names)
	return names
}

// enabledMaps returns the maps currently in the active pool.
func (a *App) enabledMaps() []domain.Map {
	out := make([]domain.Map, 0, len(a.maps))
	for _, m := range a.maps {
		if a.cfg.IsMapEnabled(m.Name) {
			out = append(out, m)
		}
	}
	return out
}
