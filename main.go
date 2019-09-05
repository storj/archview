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
	"github.com/storj/archview/graph"
)

func main() {
	log.SetFlags(0)

	format := flag.String("format", "", "format for output (dot, svg)")
	outname := flag.String("out", "", "output file")

	nocolor := flag.Bool("nocolor", false, "disable coloring")

	clustering := graph.ClusterByClass
	flag.Var(&clustering, "cluster", "clustering mode")

	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.LoadMode(^0),
	}, args...)
	if err != nil {
		log.Fatal(err)
	}

	world := arch.Analyze(pkgs...)

	if *format == "" {
		*format = strings.TrimPrefix(filepath.Ext(*outname), ".")
	}

	var out io.Writer = os.Stdout
	if *outname != "" {
		file, err := os.Create(*outname)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		out = file
	}

	switch *format {
	case "dot", "":
		_, err = (&graph.Dot{
			World:      world,
			Clustering: clustering,
			NoColor:    *nocolor,
		}).WriteTo(out)

		if err != nil {
			log.Fatalf("unable to write output: %v", err)
		}
	case "graphml":
		_, err = (&graph.GraphML{
			World: world,
		}).WriteTo(out)

		if err != nil {
			log.Fatalf("unable to write output: %v", err)
		}
	default:
		log.Fatalf("unknown format %q", *format)
	}
}
