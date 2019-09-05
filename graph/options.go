package graph

import "fmt"

type Options struct {
	Clustering Clustering
	NoColor    bool
}

type Clustering string

func (mode Clustering) String() string { return string(mode) }

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

const (
	ClusterDisabled = Clustering("")
	ClusterByClass  = Clustering("class")
)
