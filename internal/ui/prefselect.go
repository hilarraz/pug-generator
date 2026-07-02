package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// pickItem is one selectable hero or map (name + its image).
type pickItem struct {
	name string
	res  fyne.Resource
}

// pickGroup is a labelled batch of items shown together in the picker (a role,
// or a game mode).
type pickGroup struct {
	title string
	items []pickItem
}

// prefSection is a capped, image-based multi-select in the player editor. It
// shows the current picks as removable chips plus an "Ajouter" button that opens
// a card gallery capped at max. Selection is read/written through get/set so the
// same widget drives any preference slice.
type prefSection struct {
	win    fyne.Window
	title  string
	groups []pickGroup
	max    int
	get    func() []string
	set    func([]string)

	resByName map[string]fyne.Resource
	header    *widget.Label
	chips     *fyne.Container
	addBtn    *widget.Button
	root      fyne.CanvasObject
}

func newPrefSection(win fyne.Window, title string, groups []pickGroup, max int, get func() []string, set func([]string)) *prefSection {
	s := &prefSection{win: win, title: title, groups: groups, max: max, get: get, set: set}

	s.resByName = make(map[string]fyne.Resource)
	for _, g := range groups {
		for _, it := range g.items {
			s.resByName[it.name] = it.res
		}
	}

	s.header = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	s.addBtn = widget.NewButtonWithIcon("Ajouter", theme.ContentAddIcon(), s.openPicker)
	s.chips = container.NewGridWrap(fyne.NewSize(180, 40))

	head := container.NewBorder(nil, nil, s.header, s.addBtn)
	s.root = container.NewVBox(head, s.chips, widget.NewSeparator())
	s.refresh()
	return s
}

// refresh re-renders the header count, the chips, and the Ajouter state.
func (s *prefSection) refresh() {
	sel := s.get()
	s.header.SetText(fmt.Sprintf("%s  (%d/%d)", s.title, len(sel), s.max))
	if len(sel) >= s.max {
		s.addBtn.Disable()
	} else {
		s.addBtn.Enable()
	}

	objs := make([]fyne.CanvasObject, 0, len(sel))
	for _, name := range sel {
		objs = append(objs, s.chip(name))
	}
	s.chips.Objects = objs
	s.chips.Refresh()
}

// chip is a small thumbnail + name + remove button for one current pick.
func (s *prefSection) chip(name string) fyne.CanvasObject {
	img := canvas.NewImageFromResource(s.resByName[name])
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(30, 30))

	lbl := widget.NewLabel(name)
	lbl.Truncation = fyne.TextTruncateEllipsis

	del := widget.NewButtonWithIcon("", theme.CancelIcon(), func() { s.remove(name) })
	del.Importance = widget.LowImportance

	return container.NewBorder(nil, nil, img, del, lbl)
}

func (s *prefSection) selected(name string) bool {
	for _, n := range s.get() {
		if n == name {
			return true
		}
	}
	return false
}

// add includes name if there is room; it returns whether the selection changed.
func (s *prefSection) add(name string) bool {
	if s.selected(name) || len(s.get()) >= s.max {
		return false
	}
	s.set(append(s.get(), name))
	s.refresh()
	return true
}

func (s *prefSection) remove(name string) {
	cur := s.get()
	out := make([]string, 0, len(cur))
	for _, n := range cur {
		if n != name {
			out = append(out, n)
		}
	}
	s.set(out)
	s.refresh()
}

// openPicker shows the card gallery for this section, capped at max.
func (s *prefSection) openPicker() {
	count := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	updateCount := func() {
		n := len(s.get())
		msg := fmt.Sprintf("%s  (%d/%d)", s.title, n, s.max)
		if n >= s.max {
			msg += " — limite atteinte, retire un choix pour en changer"
		}
		count.SetText(msg)
	}

	body := container.NewVBox()
	for _, g := range s.groups {
		if len(g.items) == 0 {
			continue
		}
		if len(s.groups) > 1 {
			body.Add(widget.NewLabelWithStyle(strings.ToUpper(g.title), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
		}
		grid := container.NewGridWrap(cardSize)
		for _, it := range g.items {
			it := it
			var card *poolCard
			card = newPoolCard(it.name, it.res, func() {
				if s.selected(it.name) {
					s.remove(it.name)
					card.setEnabled(false)
				} else if s.add(it.name) {
					card.setEnabled(true)
				}
				updateCount()
			})
			card.setEnabled(s.selected(it.name))
			grid.Add(card)
		}
		body.Add(grid)
	}
	updateCount()

	content := container.NewBorder(
		container.NewVBox(count, widget.NewSeparator()), nil, nil, nil,
		container.NewVScroll(body),
	)
	d := dialog.NewCustom(s.title, "Fermer", content, s.win)
	d.Resize(fyne.NewSize(760, 580))
	d.SetOnClosed(s.refresh)
	d.Show()
}
