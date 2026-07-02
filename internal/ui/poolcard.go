package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// cardSize is the fixed footprint of one pool card (image above, name below).
var cardSize = fyne.NewSize(160, 150)

// poolCard is a tappable Overwatch-style card for one hero or map: its
// portrait/screenshot on top and its name underneath. Tapping toggles whether
// the item is in the active pool; an included card is lit with an orange border
// and full-strength image, an excluded one is dimmed.
type poolCard struct {
	widget.BaseWidget
	onTap func()

	bg    *canvas.Rectangle
	img   *canvas.Image
	title *widget.Label
	root  fyne.CanvasObject
}

// newPoolCard builds a card for name showing res (which may be nil, in which
// case only the placeholder background is shown). onTap fires on a click.
func newPoolCard(name string, res fyne.Resource, onTap func()) *poolCard {
	c := &poolCard{onTap: onTap}

	c.bg = canvas.NewRectangle(color.Transparent)
	c.bg.StrokeWidth = 2
	c.bg.CornerRadius = 6

	c.img = canvas.NewImageFromResource(res)
	c.img.FillMode = canvas.ImageFillContain
	c.img.SetMinSize(fyne.NewSize(140, 92))

	c.title = widget.NewLabelWithStyle(name, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	c.title.Truncation = fyne.TextTruncateEllipsis

	content := container.NewBorder(nil, c.title, nil, nil, c.img)
	c.root = container.NewStack(c.bg, container.NewPadded(content))

	c.ExtendBaseWidget(c)
	return c
}

func (c *poolCard) Tapped(*fyne.PointEvent) { c.onTap() }

func (c *poolCard) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.root)
}

// setEnabled updates the card's look to reflect whether it is in the pool.
func (c *poolCard) setEnabled(on bool) {
	if on {
		c.bg.StrokeColor = owOrange
		c.bg.FillColor = color.NRGBA{R: 0x1D, G: 0x29, B: 0x37, A: 0xFF}
		c.img.Translucency = 0
	} else {
		c.bg.StrokeColor = color.NRGBA{R: 0x2A, G: 0x3A, B: 0x4D, A: 0xFF}
		c.bg.FillColor = color.Transparent
		c.img.Translucency = 0.6
	}
	c.bg.Refresh()
	c.img.Refresh()
}
