package arch

import (
	"go/types"
	"strings"
)

// Component is a architectural piece that is related to other pieces.
type Component struct {
	Obj  types.Object
	Type types.Type

	// Links contains a list of components that it is linked to
	Links []*Link

	// Class is the functional categorization of the component.
	Class string
	// Comment has the type documentation.
	Comment string
}

// Link is a connection to another component.
type Link struct {
	Path   string
	Target *Component
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
	for _, link := range node.Links {
		names = append(names, link.Path+":"+link.Target.Name())
	}
	return node.Name() + "[" + node.Class + "] = {" + strings.Join(names, ", ") + "}"
}

// Add adds a dependency to the component.
func (node *Component) Add(link *Link) {
	node.Links = append(node.Links, link)
}

// NewLink creates a new link.
func NewLink(path string, target *Component) *Link {
	return &Link{
		Path:   strings.TrimPrefix(path, "."),
		Target: target,
	}
}
