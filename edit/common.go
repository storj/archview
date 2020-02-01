// Package edit implements different arch graph utilities.
package edit

import "github.com/storj/archview/arch"

// contains returns whether value exists in the slice.
func contains(strs []string, value string) bool {
	for _, v := range strs {
		if v == value {
			return true
		}
	}
	return false
}

// contains returns whether value exists in the slice.
func containsComponent(comps []*arch.Component, value *arch.Component) bool {
	for _, comp := range comps {
		if comp == value {
			return true
		}
	}
	return false
}
