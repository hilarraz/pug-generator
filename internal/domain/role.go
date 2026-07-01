// Package domain holds the core, dependency-free types shared across the
// application: the game entities (heroes, maps) and the participants (players).
package domain

// Role is one of Overwatch's three hero roles.
type Role string

const (
	RoleTank    Role = "Tank"
	RoleDPS     Role = "DPS"
	RoleSupport Role = "Support"
)

// Roles lists the three roles in canonical (Tank, DPS, Support) order.
var Roles = []Role{RoleTank, RoleDPS, RoleSupport}
