package archview

import (
	"go/ast"
	"go/types"
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
	Added map[types.Type]*Node
	List  []*Node
}

func NewWorld() *World {
	return &World{
		Added: map[types.Type]*Node{},
		List:  []*Node{},
	}
}

func (world *World) Empty() bool { return len(world.List) == 0 }

func (world *World) Add(n *Node) {
	world.Added[n.Type] = n
	world.List = append(world.List, n)
}

func (world *World) String() string {
	texts := []string{}
	for _, node := range world.List {
		texts = append(texts, node.String())
	}
	return strings.Join(texts, ", ")
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

	world := NewWorld()

	for _, fact := range pass.AllPackageFacts() {
		fact, ok := fact.Fact.(*Fact)
		if !ok {
			continue
		}
		for _, n := range fact.World.List {
			world.Add(n)
		}
	}

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
