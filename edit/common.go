// Package edit implements different arch graph utilities.
package edit

import "slices"

// contains returns whether value exists in the slice.
func contains(strs []string, value string) bool {
	return slices.Contains(strs, value)
}
