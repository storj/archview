package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/storj/archview"
)

func main() {
	singlechecker.Main(archview.Analyzer)
}
