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

	format := flag.String("format", "", "format for output (dot, dot-basic, svg)")
	outname := flag.String("out", "", "output file")

	var options graph.Options
	flag.StringVar(&options.TrimPrefix, "trim-prefix", "", "trim label prefix")
	flag.BoolVar(&options.NoColor, "nocolor", false, "disable coloring (dot only)")
	flag.Var(&options.Clustering, "cluster", "clustering mode (dot only)")

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
			World:   world,
			Options: options,
		}).WriteTo(out)
	case "dot-basic":
		_, err = (&graph.DotBasic{
			World:   world,
			Options: options,
		}).WriteTo(out)
	case "graphml":
		_, err = (&graph.GraphML{
			World:   world,
			Options: options,
		}).WriteTo(out)
	default:
		log.Fatalf("unknown format %q", *format)
	}
	if err != nil {
		log.Fatalf("unable to write output: %v", err)
	}
}
