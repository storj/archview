package edit

import "github.com/storj/archview/arch"

// KeepComponents only keeps the specified components in world.
func KeepComponents(world *arch.World, components []*arch.Component) {
	for _, comp := range components {
		delete(world.ByType, comp.Type)
	}

	result := world.Components[:0]
	defer func() { world.Components = result }()
	for _, comp := range world.Components {
		links := comp.Links[:0]
		for _, link := range comp.Links {
			if containsComponent(components, link.Target) {
				links = append(links, link)
			}
		}
		comp.Links = links

		if containsComponent(components, comp) {
			result = append(result, comp)
		}
	}
}
