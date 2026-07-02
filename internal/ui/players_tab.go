package ui

import (
	"fmt"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"pug-generator/internal/domain"
)

// playersTab is the roster tab. The left pane lists players with inline-editable
// names (fast rename) plus quick add / bulk-generate; the right pane edits the
// selected player's preferences.
type playersTab struct {
	root     *container.Split
	rows     *fyne.Container
	editor   *fyne.Container
	countEnt *widget.Entry
	selected int // index into a.cfg.Players, or -1
}

func (a *App) newPlayersTab() *playersTab {
	t := &playersTab{selected: -1}

	addOne := widget.NewButtonWithIcon("Ajouter", theme.ContentAddIcon(), func() {
		t.addPlayers(a, 1)
	})
	t.countEnt = widget.NewEntry()
	t.countEnt.SetText("5")
	genBtn := widget.NewButton("Générer", func() {
		if n := atoi(t.countEnt.Text); n > 0 {
			t.addPlayers(a, n)
		}
	})
	controls := container.NewBorder(nil, nil, addOne, container.NewHBox(
		widget.NewLabel("Nombre :"),
		container.NewGridWrap(fyne.NewSize(56, 36), t.countEnt),
		genBtn,
	), nil)

	t.rows = container.NewVBox()
	left := container.NewBorder(
		container.NewVBox(controls, widget.NewSeparator()), nil, nil, nil,
		container.NewVScroll(t.rows),
	)

	t.editor = container.NewVBox(t.placeholder())
	t.root = container.NewHSplit(left, container.NewVScroll(t.editor))
	t.root.SetOffset(0.42)

	t.rebuildRows(a)
	return t
}

func (t *playersTab) placeholder() fyne.CanvasObject {
	return widget.NewLabel("Sélectionne un joueur (⚙) pour éditer ses préférences.")
}

// addPlayers appends n players with placeholder names to be renamed inline.
func (t *playersTab) addPlayers(a *App, n int) {
	start := len(a.cfg.Players)
	for i := 0; i < n; i++ {
		a.cfg.Players = append(a.cfg.Players, domain.Player{
			Name:            fmt.Sprintf("Joueur %d", start+i+1),
			PreferredHeroes: map[domain.Role][]string{},
		})
	}
	t.rebuildRows(a)
}

func (t *playersTab) deletePlayer(a *App, i int) {
	if i < 0 || i >= len(a.cfg.Players) {
		return
	}
	a.cfg.Players = append(a.cfg.Players[:i], a.cfg.Players[i+1:]...)
	switch {
	case t.selected == i:
		t.selected = -1
		t.setEditor(t.placeholder())
	case t.selected > i:
		t.selected--
		t.showEditor(a) // re-bind the editor to the shifted player
	}
	t.rebuildRows(a)
}

// rebuildRows regenerates the player list. Called on structural changes only, so
// inline renaming (which doesn't rebuild) keeps focus.
func (t *playersTab) rebuildRows(a *App) {
	objs := make([]fyne.CanvasObject, 0, len(a.cfg.Players))
	for i := range a.cfg.Players {
		objs = append(objs, t.row(a, i))
	}
	if len(objs) == 0 {
		objs = append(objs, widget.NewLabel("Aucun joueur. Ajoute-en avec « Ajouter » ou « Générer »."))
	}
	t.rows.Objects = objs
	t.rows.Refresh()
}

func (t *playersTab) row(a *App, i int) fyne.CanvasObject {
	name := widget.NewEntry()
	name.SetText(a.cfg.Players[i].Name)
	name.OnChanged = func(s string) { a.cfg.Players[i].Name = s }

	prefs := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		t.selected = i
		t.showEditor(a)
	})
	del := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() { t.deletePlayer(a, i) })

	return container.NewBorder(nil, nil, nil, container.NewHBox(prefs, del), name)
}

// showEditor builds the preference editor for the selected player: image-based,
// capped multi-selects for preferred heroes (3 per role), disliked heroes (3),
// and preferred / disliked maps (3 each).
func (t *playersTab) showEditor(a *App) {
	if t.selected < 0 || t.selected >= len(a.cfg.Players) {
		return
	}
	p := &a.cfg.Players[t.selected]

	form := container.NewVBox(
		widget.NewLabelWithStyle(strings.ToUpper(p.Name), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
	)

	for _, role := range domain.Roles {
		r := role
		sec := newPrefSection(a.win, "Héros préférés — "+string(r),
			[]pickGroup{{title: string(r), items: a.heroItems(r)}}, 3,
			func() []string { return p.PreferredHeroes[r] },
			func(v []string) {
				if p.PreferredHeroes == nil {
					p.PreferredHeroes = map[domain.Role][]string{}
				}
				p.PreferredHeroes[r] = v
			},
		)
		form.Add(sec.root)
	}

	form.Add(newPrefSection(a.win, "Héros détestés", a.heroGroups(), 3,
		func() []string { return p.DislikedHeroes },
		func(v []string) { p.DislikedHeroes = v },
	).root)

	form.Add(newPrefSection(a.win, "Maps préférées", a.mapGroups(), 3,
		func() []string { return p.PreferredMaps },
		func(v []string) { p.PreferredMaps = v },
	).root)

	form.Add(newPrefSection(a.win, "Maps détestées", a.mapGroups(), 3,
		func() []string { return p.DislikedMaps },
		func(v []string) { p.DislikedMaps = v },
	).root)

	t.setEditor(form)
}

// heroItems returns the pickable cards for one role.
func (a *App) heroItems(role domain.Role) []pickItem {
	names := a.heroNamesByRole(role)
	items := make([]pickItem, 0, len(names))
	for _, n := range names {
		items = append(items, pickItem{name: n, res: heroResource(n)})
	}
	return items
}

// heroGroups returns the pickable cards for every hero, grouped by role.
func (a *App) heroGroups() []pickGroup {
	groups := make([]pickGroup, 0, len(domain.Roles))
	for _, role := range domain.Roles {
		if items := a.heroItems(role); len(items) > 0 {
			groups = append(groups, pickGroup{title: string(role), items: items})
		}
	}
	return groups
}

// mapGroups returns the pickable cards for every map, grouped by game mode.
func (a *App) mapGroups() []pickGroup {
	byMode := make(map[domain.GameMode][]string)
	for _, m := range a.maps {
		byMode[m.Mode] = append(byMode[m.Mode], m.Name)
	}
	groups := make([]pickGroup, 0, len(domain.GameModes))
	for _, mode := range domain.GameModes {
		names := byMode[mode]
		if len(names) == 0 {
			continue
		}
		sort.Strings(names)
		items := make([]pickItem, 0, len(names))
		for _, n := range names {
			items = append(items, pickItem{name: n, res: mapResource(n)})
		}
		groups = append(groups, pickGroup{title: string(mode), items: items})
	}
	return groups
}

func (t *playersTab) setEditor(content fyne.CanvasObject) {
	t.editor.Objects = []fyne.CanvasObject{content}
	t.editor.Refresh()
}

// refresh resets the tab to reflect a freshly loaded config.
func (t *playersTab) refresh(a *App) {
	t.selected = -1
	t.rebuildRows(a)
	t.setEditor(t.placeholder())
}
