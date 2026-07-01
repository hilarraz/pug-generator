package domain

// GameMode is the game mode played on a map.
type GameMode string

const (
	ModeControl    GameMode = "Control"
	ModeEscort     GameMode = "Escort"
	ModeHybrid     GameMode = "Hybrid"
	ModePush       GameMode = "Push"
	ModeFlashpoint GameMode = "Flashpoint"
	ModeClash      GameMode = "Clash"
)

// GameModes lists the game modes in a stable display order.
var GameModes = []GameMode{
	ModeControl,
	ModeEscort,
	ModeHybrid,
	ModePush,
	ModeFlashpoint,
	ModeClash,
}

// Map is an Overwatch map and its game mode.
type Map struct {
	Name string   `json:"name"`
	Mode GameMode `json:"mode"`
}
