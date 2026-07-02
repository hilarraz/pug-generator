package domain

// Player is a participant registered for PUGs, together with the preferences
// used to build and seat teams.
type Player struct {
	Name string `json:"name"`
	// PreferredHeroes maps each role to the hero names the player likes to play
	// in that role (the UI caps this at 3 per role).
	PreferredHeroes map[Role][]string `json:"preferred_heroes,omitempty"`
	// DislikedHeroes lists hero names the player would rather not play, all roles
	// combined (heroes have a fixed role, so this is a flat list; UI caps it at 3).
	DislikedHeroes []string `json:"disliked_heroes,omitempty"`
	// PreferredMaps lists map names the player enjoys (UI caps at 3).
	PreferredMaps []string `json:"preferred_maps,omitempty"`
	// DislikedMaps lists map names the player would rather avoid (UI caps at 3).
	DislikedMaps []string `json:"disliked_maps,omitempty"`
}
