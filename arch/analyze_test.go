package arch

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"golang.org/x/tools/go/packages"
)

func analyzeSource(t *testing.T, source string) *World {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "example.go", source, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	info := &types.Info{
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
		Types: make(map[ast.Expr]types.TypeAndValue),
	}
	pkg, err := (&types.Config{}).Check("example", fset, []*ast.File{file}, info)
	if err != nil {
		t.Fatal(err)
	}
	return Analyze(&packages.Package{Types: pkg, TypesInfo: info, Syntax: []*ast.File{file}})
}

func TestGroupedAnnotations(t *testing.T) {
	world := analyzeSource(t, `package example
type (
	Ignored struct{}
	// Service documents the service.
	// architecture: Service
	Service struct{}
)
// architecture: Shared
type (
	First struct{}
	// architecture: Override
	Second struct{}
)
`)
	if len(world.Components) != 3 {
		t.Fatalf("got %d components, want 3: %s", len(world.Components), world)
	}
	for i, class := range []string{"Service", "Shared", "Override"} {
		if got := world.Components[i].Class; got != class {
			t.Errorf("component %d class = %q, want %q", i, got, class)
		}
	}
}

func TestNamedTypeDependencies(t *testing.T) {
	world := analyzeSource(t, `package example
// architecture: Service
type First struct{}
// architecture: Service
type Second First
type Alias = First
// architecture: Server
type Server struct { First *Alias; Second Second }
// architecture: Factory
type Factory interface { New() *First }
// architecture: Interface
type API interface { Method() }
// architecture: Interface
type OtherAPI API
func (*First) Method() {}
var _ API = (*First)(nil)
var value = First{}
`)
	byName := map[string]*Component{}
	for _, component := range world.Components {
		byName[component.ShortName()] = component
	}
	for name, targets := range map[string][]string{
		"First": nil, "Second": nil,
		"Server": {"First", "Second"}, "Factory": {"First"},
		"API": {"First"}, "OtherAPI": nil,
	} {
		component := byName[name]
		if len(component.Links) != len(targets) {
			t.Errorf("%s: got links %v, want %v", name, component, targets)
			continue
		}
		for i, target := range targets {
			if component.Links[i].Target != byName[target] {
				t.Errorf("%s link %d = %s, want %s", name, i, component.Links[i].Target.Name(), target)
			}
		}
	}
	if !byName["API"].Links[0].Implementation {
		t.Error("interface assertion did not produce an implementation link")
	}
}
