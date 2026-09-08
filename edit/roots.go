package edit

import (
	"fmt"
	"slices"
	"strings"

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

// KeepRoots keeps roots and their transitive dependencies. Unknown roots return
// an error without changing the world. Each root is first treated as a literal
// declaration or display name, then as a comma-separated list if no literal matches.
func KeepRoots(world *arch.World, roots []string) error {
	known := make(map[string]*arch.Component, len(world.Components))
	var available []string
	for _, comp := range world.Components {
		for _, name := range []string{comp.Name(), comp.ReferenceName()} {
			if previous, exists := known[name]; exists && previous != comp {
				known[name] = nil // An ambiguous name must not silently select multiple nodes.
			} else {
				known[name] = comp
			}
		}
		available = append(available, comp.ReferenceName())
	}
	var missing []string
	var selected []*arch.Component
	for _, root := range roots {
		root = strings.TrimSpace(root)
		names := []string{root}
		if _, literal := known[root]; !literal {
			names = strings.Split(root, ",")
		}
		for _, name := range names {
			name = strings.TrimSpace(name)
			comp, exists := known[name]
			if !exists {
				missing = append(missing, name)
			} else if comp == nil {
				return fmt.Errorf("ambiguous root %q", name)
			} else {
				selected = append(selected, comp)
			}
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

	for _, comp := range selected {
		walk(comp)
	}

	keep := []*arch.Component{}
	for comp := range walked {
		keep = append(keep, comp)
	}

	KeepComponents(world, keep)
	return nil
}
