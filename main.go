package main

import (
	"flag"
	"io"
	"log"
	"os"

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

	dot := graph.Dot{
		World: world,

		GroupByClass: true,
	}

	_, _ = dot.WriteTo(out)
}
