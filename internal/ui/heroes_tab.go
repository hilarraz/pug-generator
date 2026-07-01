package ui

import "pug-generator/internal/domain"

// newHeroesPool builds the hero-pool selector, with one category per role.
func (a *App) newHeroesPool() *poolView {
	var cats []poolCategory
	for _, role := range domain.Roles {
		names := a.heroNamesByRole(role)
		if len(names) == 0 {
			continue
		}
		cats = append(cats, poolCategory{
			name:  string(role),
			color: roleColor(role),
			items: names,
		})
	}

	return newPoolView(cats,
		func(name string) bool { return a.cfg.IsHeroEnabled(name) },
		func(name string, on bool) { a.cfg.EnabledHeroes[name] = on },
	)
}
