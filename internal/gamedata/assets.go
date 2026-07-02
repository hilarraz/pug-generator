package gamedata

import (
	"embed"
	"path"
	"strings"
	"sync"

	"pug-generator/internal/slug"
)

// assets holds the hero portraits and map screenshots downloaded by cmd/gendata
// (see "Keeping game data current"). Files are named <slug.Of(name)>.<ext>.
//
//go:embed assets
var assets embed.FS

var (
	assetOnce  sync.Once
	heroImages map[string]string // slug -> path within assets
	mapImages  map[string]string
)

// HeroImage returns the embedded portrait for the named hero: its file name
// (with extension, e.g. "ana.png") and its bytes. It returns ("", nil) when no
// image is present, which callers should render as a placeholder.
func HeroImage(name string) (string, []byte) {
	assetOnce.Do(indexAssets)
	return lookupImage(heroImages, name)
}

// MapImage returns the embedded screenshot for the named map, like HeroImage.
func MapImage(name string) (string, []byte) {
	assetOnce.Do(indexAssets)
	return lookupImage(mapImages, name)
}

func lookupImage(index map[string]string, name string) (string, []byte) {
	p, ok := index[slug.Of(name)]
	if !ok {
		return "", nil
	}
	data, err := assets.ReadFile(p)
	if err != nil {
		return "", nil
	}
	return path.Base(p), data
}

func indexAssets() {
	heroImages = indexDir("assets/heroes")
	mapImages = indexDir("assets/maps")
}

// indexDir maps each file's base name (without extension) to its path, so an
// image can be found by slug regardless of whether it is a .png or .jpg.
func indexDir(dir string) map[string]string {
	index := make(map[string]string)
	entries, err := assets.ReadDir(dir)
	if err != nil {
		return index // directory missing or empty
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		base := strings.TrimSuffix(name, path.Ext(name))
		index[base] = path.Join(dir, name)
	}
	return index
}
