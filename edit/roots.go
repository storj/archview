package edit

import (
	"fmt"
	"github.com/storj/archview/arch"
	"slices"
	"strings"
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

// KeepRoots keeps roots and their transitive dependencies. Unknown roots return
// an error without changing the world.
func KeepRoots(world *arch.World, roots []string) error {
	known := make(map[string]bool, len(world.Components))
	var available []string
	for _, comp := range world.Components {
		known[comp.Name()] = true
		available = append(available, comp.Name())
	}
	var missing []string
	for _, root := range roots {
		if !known[root] {
			missing = append(missing, root)
		}
	}
	if len(missing) > 0 {
		slices.Sort(available)
		return fmt.Errorf("unknown roots %q; available components: %s", missing, strings.Join(available, ", "))
	}
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
	return nil
}
