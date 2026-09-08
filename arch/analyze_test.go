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
	world, err := Analyze(parsePackage(t, "example", source))
	if err != nil {
		t.Fatal(err)
	}
	return world
}

func parsePackage(t *testing.T, path, source string) *packages.Package {
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
	pkg, err := (&types.Config{}).Check(path, fset, []*ast.File{file}, info)
	if err != nil {
		t.Fatal(err)
	}
	return &packages.Package{ID: path, PkgPath: path, Fset: fset, Types: pkg, TypesInfo: info, Syntax: []*ast.File{file}}
}

func TestAnnotatedAliasIdentity(t *testing.T) {
	const service = "// architecture: Service\ntype Service struct{}\n"
	const alias = "// architecture: Alias\ntype Alias = Service\n"
	for _, declarations := range []string{service + alias, alias + service} {
		world := analyzeSource(t, "package example\n"+declarations+`
type Unannotated = Service
type ThroughAlias = Alias
// architecture: Pointer
type Pointer = *Service
// architecture: Cache
type Cache[T any] struct{}
// architecture: GenericAlias
type GenericAlias[T any] = Cache[T]
type UnannotatedGeneric[T any] = Cache[T]
// architecture: Server
type Server struct {
 Direct *Service
 Alias **Alias
 Plain Unannotated
 Chained ThroughAlias
 Pointer Pointer
 Generic *GenericAlias[int]
 PlainGeneric UnannotatedGeneric[string]
}
`)
		var server *Component
		for _, c := range world.Components {
			if c.ShortName() == "Server" {
				server = c
			}
		}
		want := []string{"Service", "Alias", "Service", "Alias", "Pointer", "GenericAlias", "Cache"}
		if server == nil || len(server.Links) != len(want) {
			t.Fatalf("unexpected dependencies: %s", world)
		}
		for i, target := range want {
			if got := server.Links[i].Target.ShortName(); got != target {
				t.Errorf("field %s targets %s, want %s", server.Links[i].Path, got, target)
			}
		}
	}
}

func TestImplementationAssignments(t *testing.T) {
	world := analyzeSource(t, `package example
// architecture: API
type API interface { Value() *Impl }
// architecture: Alias
type Alias = API
// architecture: Implementation
type Impl struct{}
func (*Impl) Value() *Impl { return nil }
var original API
var copied API = original
var alias Alias = original
var copiedBack API = alias
var _ API = (*Impl)(nil)
var _ API = &Impl{}
`)
	for _, c := range world.Components {
		implementations := 0
		dependencies := 0
		for _, link := range c.Links {
			if link.Implementation {
				implementations++
				if c.ShortName() != "API" || link.Target.ShortName() != "Impl" {
					t.Errorf("unexpected implementation link: %s -> %s", c.ShortName(), link.Target.ShortName())
				}
			} else {
				dependencies++
			}
		}
		want := 0
		if c.ShortName() == "API" {
			want = 1
		}
		if implementations != want {
			t.Errorf("%s has %d implementation links, want %d", c.ShortName(), implementations, want)
		}
		if c.ShortName() == "API" && dependencies != 1 {
			t.Error("implementation deduplication removed the method-result dependency")
		}
	}
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

func TestGenericAndPointerDependencies(t *testing.T) {
	world := analyzeSource(t, `package example
// architecture: Service
type Service struct{}
// architecture: Cache
type Cache[T any] struct{}
type Alias = Cache[int]
// architecture: Server
type Server struct { Cache *Alias; Service ***Service }
// architecture: Factory
type Factory interface { New() **Cache[string] }
// architecture: Interface
type API[T any] interface { Get() T }
func (*Cache[T]) Get() T { var zero T; return zero }
var _ API[int] = (*Cache[int])(nil)
type Recursive *Recursive
// architecture: Recursive
type Holder struct { Value Recursive }
`)
	want := map[string][]string{"Server": {"Cache", "Service"}, "Factory": {"Cache"}, "API": {"Cache"}, "Holder": nil}
	for _, c := range world.Components {
		targets, ok := want[c.ShortName()]
		if !ok {
			continue
		}
		if len(c.Links) != len(targets) {
			t.Fatalf("%s: want %v", c, targets)
		}
		for i, name := range targets {
			if c.Links[i].Target.ShortName() != name {
				t.Errorf("%s: link %d should target %s", c, i, name)
			}
		}
	}
}

func TestContainerAndSignatureDependencies(t *testing.T) {
	world := analyzeSource(t, `package example
// architecture: Service
type Service struct{}
type Loop []Loop
// architecture: Server
type Server struct {
 Slice []*Service
 Array [2]*Service
 Map map[*Service]chan []*Service
 Channel <-chan *Service
 Callback func(*Service) *Service
 Nested struct { Value map[string]*Service }
 Cycle Loop
}
// architecture: API
type API interface { Use(*Service); Batch(...*Service) []*Service }
`)
	for _, c := range world.Components {
		want := map[string][]string{"Server": {"Slice", "Array", "Map", "Channel", "Callback", "Nested.Value"}, "API": {"Batch", "Use"}}[c.ShortName()]
		if len(c.Links) != len(want) {
			t.Fatalf("%s: want paths %v", c, want)
		}
		for i, path := range want {
			if c.Links[i].Path != path || c.Links[i].Target.ShortName() != "Service" {
				t.Errorf("%s: link %d should be %s -> Service", c, i, path)
			}
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
