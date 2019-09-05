package graph

import "fmt"

// Options contains graph formatting and output configuration.
type Options struct {
	Clustering Clustering
	NoColor    bool
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
