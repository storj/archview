package arch

import (
	"go/types"
	"strings"
)

type World struct {
	ByType     map[types.Type]*Component
	Components []*Component
}

func NewWorld() *World {
	return &World{
		ByType:     map[types.Type]*Component{},
		Components: []*Component{},
	}
}

func (world *World) Empty() bool { return len(world.Components) == 0 }

func (world *World) Add(n *Component) {
	world.ByType[n.Type.Underlying()] = n
	world.Components = append(world.Components, n)
}

func (world *World) String() string {
	texts := []string{}
	for _, component := range world.Components {
		texts = append(texts, component.String())
	}
	return strings.Join(texts, "; ")
}
