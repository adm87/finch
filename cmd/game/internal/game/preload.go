package game

import "fmt"

// Preload loads embedded images that stay loaded for the whole session.
func (g *Game) Preload(names []string) error {
	for _, name := range names {
		if _, err := g.images.Load(g.imageFiles, name); err != nil {
			return fmt.Errorf("preload: %w", err)
		}
	}
	return nil
}
