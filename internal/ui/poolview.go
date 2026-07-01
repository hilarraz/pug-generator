package ui

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// poolCategory is one group shown in a poolView (a game mode, or a hero role).
type poolCategory struct {
	name  string
	color color.Color
	items []string
}

// poolView is a master-detail pool selector, mimicking Overwatch's menus: the
// categories sit on the left, and the selected category's items appear on the
// right as toggle "cards" that light up orange when included in the pool.
//
// It is reused for the map pool (categories = game modes) and the hero pool
// (categories = roles); the caller supplies how to read and write the enabled
// state so the same widget drives either config field.
type poolView struct {
	root fyne.CanvasObject

	isEnabled  func(name string) bool
	setEnabled func(name string, on bool)

	categories  []poolCategory
	catButtons  []*widget.Button
	itemButtons map[string]*widget.Button
	catItems    [][]*widget.Button // item buttons per category (aligned to categories)

	title    *widget.Label
	grid     *fyne.Container
	selected int
}

func newPoolView(categories []poolCategory, isEnabled func(string) bool, setEnabled func(string, bool)) *poolView {
	pv := &poolView{
		isEnabled:   isEnabled,
		setEnabled:  setEnabled,
		categories:  categories,
		itemButtons: make(map[string]*widget.Button),
		catItems:    make([][]*widget.Button, len(categories)),
	}

	// Left: one selectable button per category, with a colored swatch.
	catBox := container.NewVBox()
	for i := range categories {
		idx := i
		btn := widget.NewButton("", func() { pv.selectCategory(idx) })
		btn.Alignment = widget.ButtonAlignLeading
		pv.catButtons = append(pv.catButtons, btn)

		swatch := canvas.NewRectangle(categories[i].color)
		swatch.SetMinSize(fyne.NewSize(6, 34))
		catBox.Add(container.NewBorder(nil, nil, swatch, nil, btn))
	}

	// Right: header (title + select/clear) above the item grid.
	pv.title = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	selectAll := widget.NewButton("Tout cocher", func() { pv.setAllInCategory(true) })
	clearAll := widget.NewButton("Tout décocher", func() { pv.setAllInCategory(false) })
	header := container.NewBorder(nil, nil, pv.title, container.NewHBox(selectAll, clearAll))

	pv.grid = container.NewGridWrap(fyne.NewSize(220, 40))
	right := container.NewBorder(
		container.NewVBox(header, widget.NewSeparator()), nil, nil, nil,
		container.NewVScroll(pv.grid),
	)

	// Pre-build every item button, grouped by category.
	for i, cat := range categories {
		for _, name := range cat.items {
			name := name
			btn := widget.NewButton(name, nil)
			btn.OnTapped = func() {
				pv.setEnabled(name, !pv.isEnabled(name))
				pv.refreshItem(name)
				pv.updateCatButton(i)
			}
			pv.itemButtons[name] = btn
			pv.catItems[i] = append(pv.catItems[i], btn)
		}
	}

	split := container.NewHSplit(container.NewVScroll(catBox), right)
	split.SetOffset(0.24)
	pv.root = split

	pv.refresh()
	pv.selectCategory(0)
	return pv
}

// selectCategory shows a category's items and highlights its button.
func (pv *poolView) selectCategory(i int) {
	pv.selected = i
	for j, b := range pv.catButtons {
		if j == i {
			b.Importance = widget.HighImportance
		} else {
			b.Importance = widget.MediumImportance
		}
		b.Refresh()
	}
	pv.title.SetText(strings.ToUpper(pv.categories[i].name))

	objs := make([]fyne.CanvasObject, len(pv.catItems[i]))
	for k, b := range pv.catItems[i] {
		objs[k] = b
	}
	pv.grid.Objects = objs
	pv.grid.Refresh()
}

// setAllInCategory includes or excludes every item of the current category.
func (pv *poolView) setAllInCategory(on bool) {
	for _, name := range pv.categories[pv.selected].items {
		pv.setEnabled(name, on)
		pv.refreshItem(name)
	}
	pv.updateCatButton(pv.selected)
}

// refresh re-syncs every button with the current enabled state (after a load).
func (pv *poolView) refresh() {
	for name := range pv.itemButtons {
		pv.refreshItem(name)
	}
	for i := range pv.categories {
		pv.updateCatButton(i)
	}
}

func (pv *poolView) refreshItem(name string) {
	b := pv.itemButtons[name]
	if pv.isEnabled(name) {
		b.Importance = widget.HighImportance
	} else {
		b.Importance = widget.MediumImportance
	}
	b.Refresh()
}

func (pv *poolView) updateCatButton(i int) {
	cat := pv.categories[i]
	enabled := 0
	for _, name := range cat.items {
		if pv.isEnabled(name) {
			enabled++
		}
	}
	pv.catButtons[i].SetText(fmt.Sprintf("%s   %d/%d", strings.ToUpper(cat.name), enabled, len(cat.items)))
}
