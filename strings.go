package main

import "strings"

// Strings is a flag for collecting multiple strings.
type Strings []string

// Empty returns whether slice contains values.
func (strs Strings) Empty() bool { return len(strs) == 0 }

// String returns the string representation of the clustering.
func (strs Strings) String() string { return strings.Join(strs, ",") }

// Set implements flag.Value such that Strings can be used as a flag.
func (strs *Strings) Set(value string) error {
	*strs = append(*strs, strings.Split(value, ",")...)
	return nil
}
