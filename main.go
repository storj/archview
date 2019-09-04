package main

import (
	"flag"
	"log"
	"os"

	"golang.org/x/tools/go/packages"

	"github.com/storj/archview/arch"
	"github.com/storj/archview/graph"
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	pkgs, err := packages.Load(&packages.Config{
		Mode: NeedAll,
	}, args...)
	if err != nil {
		log.Fatal(err)
	}

	world := arch.Analyze(pkgs...)

	dot := graph.Dot{
		World: world,

		GroupByClass: true,
	}

	_, _ = dot.WriteTo(os.Stdout)
}

const NeedAll = packages.NeedName |
	packages.NeedFiles |
	packages.NeedCompiledGoFiles |
	packages.NeedImports |
	packages.NeedDeps |
	packages.NeedExportsFile |
	packages.NeedTypes |
	packages.NeedSyntax |
	packages.NeedTypesInfo |
	packages.NeedTypesSizes
