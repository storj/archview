package edit

import "github.com/storj/archview/arch"

// KeepComponents only keeps the specified components in world.
func KeepComponents(world *arch.World, components []*arch.Component) {
	keep := make(map[*arch.Component]bool, len(components))
	for _, comp := range components {
		keep[comp] = true
	}
	for typ, comp := range world.ByType {
		if !keep[comp] {
			delete(world.ByType, typ)
		}
	}

	result := world.Components[:0]
	defer func() { world.Components = result }()
	for _, comp := range world.Components {
		links := comp.Links[:0]
		for _, link := range comp.Links {
			if keep[link.Target] {
				links = append(links, link)
			}
		}
		comp.Links = links

		if keep[comp] {
			result = append(result, comp)
		}
	}
}
