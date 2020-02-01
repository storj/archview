package edit

import (
	"github.com/storj/archview/arch"
)

// ComponentsWithClass returns list of components with the  class.
func ComponentsWithClass(graph *arch.World, classes []string) []*arch.Component {
	var components []*arch.Component
	for _, component := range graph.Components {
		if contains(classes, component.Class) {
			components = append(components, component)
		}
	}
	return components
}

// RemoveClasses removes any components that has class in the specified list.
func RemoveClasses(graph *arch.World, classes []string) {
	var keep []*arch.Component
	for _, component := range graph.Components {
		if !contains(classes, component.Class) {
			keep = append(keep, component)
		}
	}

	KeepComponents(graph, keep)
}
