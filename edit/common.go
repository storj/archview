// Package edit implements different arch graph utilities.
package edit

import "slices"

import "github.com/storj/archview/arch"

// contains returns whether value exists in the slice.
func contains(strs []string, value string) bool {
	return slices.Contains(strs, value)
}

// contains returns whether value exists in the slice.
func containsComponent(comps []*arch.Component, value *arch.Component) bool {
	return slices.Contains(comps, value)
}
