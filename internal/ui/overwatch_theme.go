package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	"pug-generator/internal/domain"
)

// owOrange is the signature Overwatch accent color.
var owOrange = color.NRGBA{R: 0xFA, G: 0x9C, B: 0x1E, A: 0xFF}

// overwatchTheme is a dark, Overwatch-flavored theme: near-black navy background
// with an orange accent. It delegates anything it doesn't override to Fyne's
// default dark theme.
type overwatchTheme struct{}

var _ fyne.Theme = overwatchTheme{}

func (overwatchTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 0x10, G: 0x18, B: 0x22, A: 0xFF}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 0xEC, G: 0xF0, B: 0xF4, A: 0xFF}
	case theme.ColorNamePrimary:
		return owOrange
	case theme.ColorNameButton:
		return color.NRGBA{R: 0x1D, G: 0x29, B: 0x37, A: 0xFF}
	case theme.ColorNameHover:
		return color.NRGBA{R: 0x2A, G: 0x3A, B: 0x4D, A: 0xFF}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 0x18, G: 0x22, B: 0x2E, A: 0xFF}
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 0x2A, G: 0x3A, B: 0x4D, A: 0xFF}
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 0x76, G: 0x86, B: 0x96, A: 0xFF}
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 0x55, G: 0x63, B: 0x70, A: 0xFF}
	case theme.ColorNameSelection:
		return color.NRGBA{R: 0xFA, G: 0x9C, B: 0x1E, A: 0x55}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x66}
	}
	return theme.DefaultTheme().Color(name, theme.VariantDark)
}

func (overwatchTheme) Font(s fyne.TextStyle) fyne.Resource     { return theme.DefaultTheme().Font(s) }
func (overwatchTheme) Icon(n fyne.ThemeIconName) fyne.Resource { return theme.DefaultTheme().Icon(n) }
func (overwatchTheme) Size(n fyne.ThemeSizeName) float32       { return theme.DefaultTheme().Size(n) }

// roleColor is the canonical Overwatch color for a role.
func roleColor(r domain.Role) color.Color {
	switch r {
	case domain.RoleTank:
		return color.NRGBA{R: 0x5A, G: 0x9B, B: 0xD4, A: 0xFF} // blue
	case domain.RoleDPS:
		return color.NRGBA{R: 0xE0, G: 0x5A, B: 0x48, A: 0xFF} // red
	case domain.RoleSupport:
		return color.NRGBA{R: 0x85, G: 0xBB, B: 0x65, A: 0xFF} // green
	}
	return owOrange
}

// modeColor is a distinct accent color per game mode, for the category swatches.
func modeColor(m domain.GameMode) color.Color {
	switch m {
	case domain.ModeControl:
		return color.NRGBA{R: 0x6D, G: 0xA6, B: 0xD6, A: 0xFF}
	case domain.ModeEscort:
		return color.NRGBA{R: 0xE8, G: 0x91, B: 0x3C, A: 0xFF}
	case domain.ModeHybrid:
		return color.NRGBA{R: 0x9B, G: 0x6D, B: 0xC6, A: 0xFF}
	case domain.ModePush:
		return color.NRGBA{R: 0x46, G: 0xB5, B: 0xA5, A: 0xFF}
	case domain.ModeFlashpoint:
		return color.NRGBA{R: 0xE0, G: 0xC2, B: 0x4E, A: 0xFF}
	case domain.ModeClash:
		return color.NRGBA{R: 0xD4, G: 0x6A, B: 0x9F, A: 0xFF}
	}
	return owOrange
}
