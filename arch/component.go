package arch

import (
	"go/types"
	"strings"
)

// Component is a architectural piece that is related to other pieces.
type Component struct {
	Obj  types.Object
	Type types.Type

	// Deps contains a list of components that it depends on.
	Deps []*Dep

	// Class is the functional categorization of the component.
	Class string
	// Comment has the type documentation.
	Comment string
}

// Dep is a connection to another component.
type Dep struct {
	Path string
	Dep  *Component
}

// Name returns the fully qualified name of the component.
func (node *Component) Name() string {
	return node.Type.String()
}

// Package returns the package name.
func (node *Component) Package() string {
	return node.Obj.Pkg().Path()
}

// ShortName returns name without package.
func (node *Component) ShortName() string {
	return node.Obj.Name()
}

// String creates a short description of the component that is useful for debugging.
func (node *Component) String() string {
	names := []string{}
	for _, dep := range node.Deps {
		names = append(names, dep.Path+":"+dep.Dep.Name())
	}
	return node.Name() + "[" + node.Class + "] = {" + strings.Join(names, ", ") + "}"
}

// Add adds a dependency to the component.
func (node *Component) Add(dep *Dep) {
	node.Deps = append(node.Deps, dep)
}

// NewDep creates a new dependency.
func NewDep(path string, dep *Component) *Dep {
	return &Dep{
		Path: strings.TrimPrefix(path, "."),
		Dep:  dep,
	}
}
