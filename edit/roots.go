package edit

import (
	"github.com/storj/archview/arch"
)

// Roots returns components without incoming links.
func Roots(world *arch.World) []*arch.Component {
	parents := map[*arch.Component]int{}
	for _, comp := range world.Components {
		parents[comp] += 0
		for _, link := range comp.Links {
			parents[link.Target]++
		}
	}

	var roots []*arch.Component
	for comp, count := range parents {
		if count == 0 {
			roots = append(roots, comp)
		}
	}
	return roots
}

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
