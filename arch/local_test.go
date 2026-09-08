package arch

import (
	"strings"
	"testing"
)

func TestLocalAnnotationsRejected(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		errors       int
	}{
		{"same name in different functions", `
func first() {
 // architecture: First
 type Node struct{}
}
func second() {
 // architecture: Second
 type Node struct{}
}`, 2},
		{"method alias", `
type Owner struct{}
func (Owner) build() {
 // architecture: Local
 type Node = int
}`, 1},
		{"closure", `
var build = func() {
 // architecture: Local
 type Node struct{}
}`, 1},
		{"nested block", `
func build() {
 if true {
  // architecture: Local
  type Node struct{}
 }
}`, 1},
		{"grouped declaration", `
func build() {
 // architecture: Local
 type (
  First struct{}
  Second struct{}
 )
}`, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			world, err := Analyze(parsePackage(t, "example", "package example\n"+tc.source))
			if err == nil || world != nil {
				t.Fatalf("expected local annotation error, got %v, %v", world, err)
			}
			for _, text := range []string{"example.go:", "architecture annotation on local type", "move the type to package scope"} {
				if got := strings.Count(err.Error(), text); got != tc.errors {
					t.Errorf("got %d occurrences of %q, want %d: %v", got, text, tc.errors, err)
				}
			}
		})
	}
}

func TestUnannotatedLocalTypesAllowed(t *testing.T) {
	world := analyzeSource(t, `package example
// architecture: Service
type Service struct{}
func first() { type Node struct { Service *Service } }
func second() { type Node = Service }
`)
	if len(world.Components) != 1 || world.Components[0].ShortName() != "Service" {
		t.Fatalf("unannotated local types changed components: %s", world)
	}
}
