package archview

import (
	"fmt"
	"go/ast"
	"go/types"
	"os"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name: "archview",
	Doc:  "creates an architecutral view",
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	FactTypes: []analysis.Fact{&Fact{}},
}

type Fact struct {
	World *World
}

func (*Fact) AFact() {}

func (fact *Fact) String() string {
	return fact.World.String()
}

type World struct {
	ByType map[types.Type]*Node
	List   []*Node
}

func NewWorld() *World {
	return &World{
		ByType: map[types.Type]*Node{},
		List:   []*Node{},
	}
}

func (world *World) Empty() bool { return len(world.List) == 0 }

func (world *World) Add(n *Node) {
	world.ByType[n.Type.Underlying()] = n
	world.List = append(world.List, n)
}

func (world *World) String() string {
	texts := []string{}
	for _, node := range world.List {
		texts = append(texts, node.String())
	}
	return strings.Join(texts, "; ")
}

type Node struct {
	Type  types.Type
	Class string
	Deps  []*Dep
}

func (node *Node) Name() string {
	return node.Type.String()
}

func (node *Node) String() string {
	names := []string{}
	for _, dep := range node.Deps {
		names = append(names, dep.Path+":"+dep.Node.Name())
	}
	return node.Name() + "[" + node.Class + "] = {" + strings.Join(names, ", ") + "}"
}

type Dep struct {
	Path string
	Node *Node
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	local := NewWorld()

	inspect.Preorder([]ast.Node{
		(*ast.TypeSpec)(nil),
	}, func(n ast.Node) {
		ts := n.(*ast.TypeSpec)

		tag, ok := ExtractTypeTag(ts)
		if !ok {
			return
		}

		obj := pass.TypesInfo.Defs[ts.Name]
		local.Add(&Node{
			Type:  obj.Type(),
			Class: tag,
		})
	})

	world := NewWorld()
	include := func(local *World) {
		for _, n := range local.List {
			world.Add(n)
		}
	}

	for _, fact := range pass.AllPackageFacts() {
		fact, ok := fact.Fact.(*Fact)
		if !ok {
			continue
		}
		include(fact.World)
	}
	include(local)

	for _, node := range local.List {
		node.Deps = FindDeps(pass, "", world, node, node.Type)
	}

	if !local.Empty() {
		pass.ExportPackageFact(&Fact{local})
	}

	return nil, nil
}

func ExtractTypeTag(ts *ast.TypeSpec) (tag string, ok bool) {
	if ts.Comment == nil {
		return "", false
	}

	for _, c := range ts.Comment.List {
		if strings.HasPrefix(c.Text, "// archview:") {
			tag := strings.TrimPrefix(c.Text, "// archview:")
			return strings.TrimSpace(tag), true
		}
	}

	return "", false
}

func FindDeps(pass *analysis.Pass, path string, world *World, node *Node, typ types.Type) (deps []*Dep) {
	switch t := typ.Underlying().(type) {
	case *types.Interface:
		for i := 0; i < t.NumMethods(); i++ {
			method := t.Method(i)
			switch m := method.Type().(type) {
			case *types.Signature:
				result := m.Results()
				for i := 0; i < result.Len(); i++ {
					at := result.At(i)

					dep := world.ByType[at.Type().Underlying()]
					if dep != nil {
						deps = append(deps, &Dep{
							Path: strings.TrimPrefix(path+"."+method.Name(), "."),
							Node: dep,
						})
					}
				}
			default:
				fmt.Fprintf(os.Stderr, "unhandled method type %T\n", m)
			}
		}

	case *types.Struct:
		for i := 0; i < t.NumFields(); i++ {
			field := t.Field(i)
			underlying := tryDeref(field.Type().Underlying())

			dep := world.ByType[underlying]
			if dep != nil {
				deps = append(deps, &Dep{
					Path: strings.TrimPrefix(path+"."+field.Name(), "."),
					Node: dep,
				})
				continue
			}

			switch f := underlying.(type) {
			case *types.Pointer:
				deps = append(deps, FindDeps(pass, path+"."+field.Name(), world, node, f)...)
			case *types.Struct:
				deps = append(deps, FindDeps(pass, path+"."+field.Name(), world, node, f)...)
			default:
				fmt.Fprintf(os.Stderr, "unhandled method %q type %T\n", path, f)
			}
		}

	default:
		fmt.Fprintf(os.Stderr, "unhandled type %T\n", t)
	}
	return deps
}

func tryDeref(t types.Type) types.Type {
	if ptr, ok := t.(*types.Pointer); ok {
		return ptr.Elem().Underlying()
	}
	return t.Underlying()
}
