package domain

// Hero is a playable Overwatch hero and the role it belongs to.
type Hero struct {
	Name string `json:"name"`
	Role Role   `json:"role"`
}
