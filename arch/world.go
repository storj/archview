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
	world.ByType[n.Type] = n
	world.Components = append(world.Components, n)
}

// componentFor preserves annotated aliases and resolves unannotated aliases,
// pointers, and generic instances to their component declarations.
func (world *World) componentFor(typ types.Type) *Component {
	for {
		if component := world.ByType[typ]; component != nil {
			return component
		}
		switch t := typ.(type) {
		case *types.Alias:
			if component := world.ByType[t.Origin()]; component != nil {
				return component
			}
			typ = t.Rhs()
		case *types.Pointer:
			typ = t.Elem()
		case *types.Named:
			return world.ByType[t.Origin()]
		default:
			return nil
		}
	}
}

// String creates a short description of the world that is useful for debugging.
func (world *World) String() string {
	texts := []string{}
	for _, component := range world.Components {
		texts = append(texts, component.String())
	}
	return strings.Join(texts, "; ")
}
