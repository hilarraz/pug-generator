package domain

// Player is a participant registered for PUGs, together with the preferences
// used to build and seat teams.
type Player struct {
	Name string `json:"name"`
	// PreferredHeroes maps each role to the hero names the player likes to play
	// in that role.
	PreferredHeroes map[Role][]string `json:"preferred_heroes,omitempty"`
	// PreferredMaps lists map names the player enjoys.
	PreferredMaps []string `json:"preferred_maps,omitempty"`
	// DislikedMaps lists map names the player would rather avoid.
	DislikedMaps []string `json:"disliked_maps,omitempty"`
}
