package ui

import (
	"fmt"
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

// showEditor builds the preference editor for the selected player.
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
		grp := widget.NewCheckGroup(a.heroNamesByRole(role), nil)
		grp.SetSelected(p.PreferredHeroes[role])
		r := role
		grp.OnChanged = func(sel []string) {
			if p.PreferredHeroes == nil {
				p.PreferredHeroes = map[domain.Role][]string{}
			}
			p.PreferredHeroes[r] = sel
		}
		form.Add(widget.NewCard("Héros préférés — "+string(role), "", grp))
	}

	mapNames := a.mapNames()

	prefMaps := widget.NewCheckGroup(mapNames, nil)
	prefMaps.SetSelected(p.PreferredMaps)
	prefMaps.OnChanged = func(sel []string) { p.PreferredMaps = sel }
	form.Add(widget.NewCard("Maps préférées", "", prefMaps))

	disMaps := widget.NewCheckGroup(mapNames, nil)
	disMaps.SetSelected(p.DislikedMaps)
	disMaps.OnChanged = func(sel []string) { p.DislikedMaps = sel }
	form.Add(widget.NewCard("Maps détestées", "", disMaps))

	t.setEditor(form)
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
