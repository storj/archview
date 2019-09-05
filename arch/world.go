package arch

import (
	"go/types"
	"strings"
)

// World contains the component and relation information.
type World struct {
	ByType     map[types.Type]*Component
	Components []*Component
}

// NewWorld creates a new world.
func NewWorld() *World {
	return &World{
		ByType:     map[types.Type]*Component{},
		Components: []*Component{},
	}
}

// Add adds a new component to world.
func (world *World) Add(n *Component) {
	world.ByType[n.Type.Underlying()] = n
	world.Components = append(world.Components, n)
}

// String creates a short description of the world that is useful for debugging.
func (world *World) String() string {
	texts := []string{}
	for _, component := range world.Components {
		texts = append(texts, component.String())
	}
	return strings.Join(texts, "; ")
}
