// Package slug turns a hero or map name into a filesystem-safe identifier used
// to name (and later look up) its embedded image.
//
// It is deliberately a tiny, dependency-light leaf package so that both
// cmd/gendata (which names the downloaded image files) and internal/gamedata
// (which looks them up) can share the exact same rule without either importing
// the other's go:embed.
package slug

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Of returns the slug for a name: accents folded, lowercased, and reduced to
// [a-z0-9]. For example "D.Va" -> "dva", "Lúcio" -> "lucio",
// "Wrecking Ball" -> "wreckingball". The writer (cmd/gendata) and the reader
// (internal/gamedata) must both call this so their names agree.
func Of(name string) string {
	var b strings.Builder
	for _, r := range norm.NFD.String(name) {
		if unicode.Is(unicode.Mn, r) {
			continue // combining mark left over from decomposing an accent
		}
		r = unicode.ToLower(r)
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}
