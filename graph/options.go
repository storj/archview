package graph

import (
	"fmt"
	"strings"

	"github.com/storj/archview/arch"
)

// Options contains graph formatting and output configuration.
type Options struct {
	TrimPrefix string
	Clustering Clustering
	NoColor    bool

	SkipClasses Strings
}

// Skip returns whether component should be skipped in output.
func (opts *Options) Skip(component *arch.Component) bool {
	for _, class := range opts.SkipClasses {
		if class == component.Class {
			return true
		}
	}
	return false
}

// Clustering is an enum for clustering modes.
type Clustering string

// Clustering lists different ways of clustering.
const (
	ClusterDisabled = Clustering("")
	ClusterByClass  = Clustering("class")
)

// String returns the string representation of the clustering.
func (mode Clustering) String() string { return string(mode) }

// Set implements flag.Value such that Clustering can be used as a flag.
func (mode *Clustering) Set(value string) error {
	switch Clustering(value) {
	case ClusterDisabled:
		*mode = ClusterDisabled
	case ClusterByClass:
		*mode = ClusterByClass
	default:
		return fmt.Errorf("unknown clustering mode %q", value)
	}
	return nil
}

// Strings is a flag for collecting multiple strings.
type Strings []string

// String returns the string representation of the clustering.
func (strs Strings) String() string { return strings.Join(strs, ",") }

// Set implements flag.Value such that Strings can be used as a flag.
func (strs *Strings) Set(value string) error {
	*strs = append(*strs, strings.Split(value, ",")...)
	return nil
}
