package main

import (
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/storj/archview/arch"
	"github.com/storj/archview/edit"
	"github.com/storj/archview/graph"
)

func main() {
	log.SetFlags(0)

	format := flag.String("format", "", "format for output (dot, dot-basic, graphml, elk)")
	outname := flag.String("out", "", "output file")

	var skipClasses Strings
	var roots Strings
	flag.Var(&skipClasses, "skip-class", "skip components with the specified class in output")
	flag.Var(&roots, "root", "only display components that are dependencies of root")

	var options graph.Options
	options.Clustering = graph.ClusterByClass
	flag.StringVar(&options.TrimPrefix, "trim-prefix", "", "trim label prefix")
	flag.BoolVar(&options.NoColor, "nocolor", false, "disable coloring (dot only)")
	flag.Var(&options.Clustering, "cluster", "clustering mode (dot only)")

	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}
	if *format == "" {
		*format = strings.TrimPrefix(filepath.Ext(*outname), ".")
	}
	switch *format {
	case "", "dot", "dot-basic", "graphml", "elk":
	default:
		log.Fatalf("unknown format %q", *format)
	}

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.LoadMode(^0),
	}, args...)
	if err != nil {
		log.Fatal(err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}
	if len(pkgs) == 0 {
		log.Fatal("no packages matched")
	}

	world, err := arch.Analyze(pkgs...)
	if err != nil {
		log.Fatal(err)
	}

	if !roots.Empty() {
		if err := edit.KeepRoots(world, roots); err != nil {
			log.Fatal(err)
		}
	}
	if !skipClasses.Empty() {
		edit.RemoveClasses(world, skipClasses)
	}

	var renderer io.WriterTo
	switch *format {
	case "dot", "":
		renderer = &graph.Dot{
			World:   world,
			Options: options,
		}
	case "dot-basic":
		renderer = &graph.DotBasic{
			World:   world,
			Options: options,
		}
	case "graphml":
		renderer = &graph.GraphML{
			World:   world,
			Options: options,
		}
	case "elk":
		renderer = &graph.ELK{
			World:   world,
			Options: options,
		}
	default:
		log.Fatalf("unknown format %q", *format)
	}
	if *outname == "" {
		_, err = renderer.WriteTo(os.Stdout)
	} else {
		err = writeOutput(*outname, renderer)
	}
	if err != nil {
		log.Fatalf("unable to write output: %v", err)
	}
}
