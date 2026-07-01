package ui

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"pug-generator/internal/domain"
	"pug-generator/internal/generator"
)

// pugTab drives team generation: the user picks a format (team count, role
// queue, composition) and generates teams from the enabled pools and roster.
type pugTab struct {
	content fyne.CanvasObject

	teamCount   *widget.Select
	roleQueue   *widget.Check
	compTank    *widget.Select
	compDPS     *widget.Select
	compSupport *widget.Select
	teamSize    *widget.Select

	results *fyne.Container
}

func (a *App) newPugTab() *pugTab {
	t := &pugTab{}
	counts := []string{"0", "1", "2", "3", "4"}

	t.teamCount = widget.NewSelect([]string{"2", "3", "4"}, nil)
	t.teamCount.SetSelected("2")

	t.compTank = newCountSelect(counts, "1")
	t.compDPS = newCountSelect(counts, "2")
	t.compSupport = newCountSelect(counts, "2")
	t.teamSize = newCountSelect([]string{"1", "2", "3", "4", "5", "6"}, "5")

	t.roleQueue = widget.NewCheck("Role queue", func(on bool) { t.updateEnabled(on) })
	t.roleQueue.SetChecked(true)

	genBtn := widget.NewButtonWithIcon("Générer", theme.MediaPlayIcon(), func() { t.generate(a) })
	genBtn.Importance = widget.HighImportance

	controls := container.NewVBox(
		container.NewHBox(widget.NewLabel("Équipes :"), t.teamCount, t.roleQueue),
		container.NewHBox(
			widget.NewLabel("Composition —  Tank :"), t.compTank,
			widget.NewLabel("DPS :"), t.compDPS,
			widget.NewLabel("Support :"), t.compSupport,
		),
		container.NewHBox(widget.NewLabel("Taille d'équipe (open queue) :"), t.teamSize),
		genBtn,
		widget.NewSeparator(),
	)

	t.results = container.NewVBox(widget.NewLabel("Configure le format puis clique sur Générer."))
	t.content = container.NewBorder(controls, nil, nil, nil, container.NewVScroll(t.results))
	return t
}

func newCountSelect(options []string, selected string) *widget.Select {
	s := widget.NewSelect(options, nil)
	s.SetSelected(selected)
	return s
}

// updateEnabled enables the controls relevant to the current queue mode.
func (t *pugTab) updateEnabled(roleQueue bool) {
	toggle := func(w *widget.Select, on bool) {
		if on {
			w.Enable()
		} else {
			w.Disable()
		}
	}
	toggle(t.compTank, roleQueue)
	toggle(t.compDPS, roleQueue)
	toggle(t.compSupport, roleQueue)
	toggle(t.teamSize, !roleQueue)
}

func (t *pugTab) generate(a *App) {
	opts := generator.Options{
		TeamCount: atoi(t.teamCount.Selected),
		RoleQueue: t.roleQueue.Checked,
	}
	if opts.RoleQueue {
		opts.Composition = map[domain.Role]int{
			domain.RoleTank:    atoi(t.compTank.Selected),
			domain.RoleDPS:     atoi(t.compDPS.Selected),
			domain.RoleSupport: atoi(t.compSupport.Selected),
		}
	} else {
		opts.TeamSize = atoi(t.teamSize.Selected)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	res, err := generator.Generate(rng, a.cfg.Players, a.enabledMaps(), opts)
	if err != nil {
		dialog.ShowError(err, a.win)
		return
	}
	t.showResult(res)
}

func (t *pugTab) showResult(res *generator.Result) {
	items := make([]fyne.CanvasObject, 0)

	if res.Map != nil {
		items = append(items, widget.NewCard("Map", string(res.Map.Mode)+" — "+res.Map.Name, nil))
	}

	teamCards := make([]fyne.CanvasObject, 0, len(res.Teams))
	for i, team := range res.Teams {
		lines := container.NewVBox()
		for _, a := range team.Players {
			label := a.Player.Name
			if a.Role != "" {
				label = string(a.Role) + " — " + a.Player.Name
			}
			lines.Add(widget.NewLabel(label))
		}
		teamCards = append(teamCards, widget.NewCard(fmt.Sprintf("Équipe %d", i+1), "", lines))
	}
	items = append(items, container.NewGridWithColumns(len(teamCards), teamCards...))

	if len(res.Bench) > 0 {
		bench := container.NewVBox()
		for _, p := range res.Bench {
			bench.Add(widget.NewLabel(p.Name))
		}
		items = append(items, widget.NewCard("Remplaçants", "", bench))
	}

	t.results.Objects = items
	t.results.Refresh()
}

// atoi parses a value coming from a controlled Select; it defaults to 0.
func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
