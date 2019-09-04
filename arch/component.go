package arch

import (
	"go/types"
	"strings"
)

type Component struct {
	Obj   types.Object
	Type  types.Type
	Class string
	Deps  []*Dep
}

type Dep struct {
	Path string
	Dep  *Component
}

func (node *Component) Name() string {
	return node.Type.String()
}

func (node *Component) String() string {
	names := []string{}
	for _, dep := range node.Deps {
		names = append(names, dep.Path+":"+dep.Dep.Name())
	}
	return node.Name() + "[" + node.Class + "] = {" + strings.Join(names, ", ") + "}"
}

func (node *Component) Add(path string, dep *Component) {
	node.Deps = append(node.Deps, &Dep{
		Path: strings.TrimPrefix(path, "."),
		Dep:  dep,
	})
}
