package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// settingsTab is the Paramètres tab: session-wide generation options bound to
// the config.
type settingsTab struct {
	root   fyne.CanvasObject
	assign *widget.Check
}

func (a *App) newSettingsTab() *settingsTab {
	t := &settingsTab{}

	t.assign = widget.NewCheck("Assigner un héros à chaque joueur (plutôt qu'un simple rôle)", func(on bool) {
		a.cfg.Settings.AssignHeroes = on
	})
	t.assign.SetChecked(a.cfg.Settings.AssignHeroes)

	desc := widget.NewLabel("Coché : à la génération, chaque joueur reçoit un héros tiré de ses préférences " +
		"pour le rôle attribué, en évitant ses héros détestés et les doublons dans une même équipe.\n" +
		"Décoché : seul un rôle est attribué (comportement par défaut).")
	desc.Wrapping = fyne.TextWrapWord

	t.root = container.NewVBox(
		widget.NewCard("Génération", "", container.NewVBox(t.assign, desc)),
	)
	return t
}

// refresh re-syncs the controls with the current config (after a load).
func (t *settingsTab) refresh(a *App) {
	t.assign.SetChecked(a.cfg.Settings.AssignHeroes)
}
