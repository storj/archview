package archview_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/storj/archview"
)

func TestFromFileSystem(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, archview.Analyzer, "example/...")
}
