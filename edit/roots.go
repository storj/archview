package edit

import (
	"github.com/storj/archview/arch"
)

// KeepRoots keeps any component that is a dependency to root.
func KeepRoots(world *arch.World, roots []string) {
	walked := map[*arch.Component]bool{}
	var walk func(*arch.Component)
	walk = func(comp *arch.Component) {
		if walked[comp] {
			return
		}
		walked[comp] = true

		for _, link := range comp.Links {
			walk(link.Target)
		}
	}

	for _, comp := range world.Components {
		if contains(roots, comp.Name()) {
			walk(comp)
		}
	}

	keep := []*arch.Component{}
	for comp := range walked {
		keep = append(keep, comp)
	}

	KeepComponents(world, keep)
}
