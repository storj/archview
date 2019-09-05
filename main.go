package main

import (
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"

	"github.com/storj/archview/arch"
	"github.com/storj/archview/graph"
)

func main() {
	log.SetFlags(0)

	outname := flag.String("out", "", "output file")
	flag.Parse()

	var out io.Writer = os.Stdout
	if *outname != "" {
		file, err := os.Create(*outname)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		out = file
	}

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

	var format io.WriterTo
	switch ext := filepath.Ext(*outname); ext {
	case ".dot", "":
		format = &graph.Dot{
			World:        world,
			GroupByClass: false,
		}
	case ".graphml":
		format = &graph.GraphML{
			World: world,
		}
	default:
		log.Fatalf("unknown format %q", ext)
	}

	_, err = format.WriteTo(out)
	if err != nil {
		log.Fatal(err)
	}
}
