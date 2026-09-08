package graph

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os/exec"
	"strings"
	"testing"

	"github.com/storj/archview/arch"
	"github.com/storj/archview/graph/graphml"
)

func component(pkgPath, name string) *arch.Component {
	pkg := types.NewPackage(pkgPath, "example")
	obj := types.NewTypeName(token.NoPos, pkg, name, nil)
	typ := types.NewNamed(obj, types.NewStruct(nil, nil), nil)
	return &arch.Component{Obj: obj, Type: typ, Comment: "Documentation <&>"}
}

func TestDotGenericConstraintLabel(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "example.go", `package example
type Cache[T interface{ ~int | ~string }] struct{}
`, 0)
	if err != nil {
		t.Fatal(err)
	}
	pkg, err := (&types.Config{}).Check("example", fset, []*ast.File{file}, nil)
	if err != nil {
		t.Fatal(err)
	}
	obj := pkg.Scope().Lookup("Cache")
	world := arch.NewWorld()
	world.Add(&arch.Component{Obj: obj, Type: obj.Type()})
	for _, noColor := range []bool{false, true} {
		t.Run(fmt.Sprint(noColor), func(t *testing.T) {
			var out bytes.Buffer
			if _, err := (&Dot{World: world, Options: Options{NoColor: noColor}}).WriteTo(&out); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out.String(), "shape=box") || !strings.Contains(out.String(), fmt.Sprintf("label=%q", obj.Type().String())) {
				t.Fatalf("expected plain shape and complete type label:\n%s", &out)
			}
			if dot, err := exec.LookPath("dot"); err == nil {
				cmd := exec.Command(dot, "-Tsvg")
				cmd.Stdin = &out
				if result, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("Graphviz rejected generic label: %v\n%s", err, result)
				}
			}
		})
	}
}

func TestDotBasicPreservesNodes(t *testing.T) {
	a := component("example/p", "A")
	b := component("p", "A")
	isolated := component("p", "Isolated")
	a.Add(arch.NewLink("field", b))
	world := &arch.World{Components: []*arch.Component{a, b, isolated}}
	var out bytes.Buffer
	_, err := (&DotBasic{World: world, Options: Options{TrimPrefix: "example/"}}).WriteTo(&out)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"digraph G {", `"example/p.A" [label="p.A"];`,
		`"example/p.A" -> "p.A";`, `"p.Isolated";`,
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output missing %q:\n%s", want, &out)
		}
	}
	if dot, err := exec.LookPath("dot"); err == nil {
		cmd := exec.Command(dot, "-Tcanon")
		cmd.Stdin = bytes.NewReader(out.Bytes())
		if result, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Graphviz rejected DOT: %v\n%s", err, result)
		}
	}
}

func TestGraphMLKeysAndIdentity(t *testing.T) {
	a := component("example/a-b", "Node")
	b := component("example/a_b", "Node")
	a.Add(arch.NewLink("field <&>", b))
	a.Add(&arch.Link{Target: b, Explicit: true, Label: "sends <requests> & events"})
	world := &arch.World{Components: []*arch.Component{a, b}}
	var out bytes.Buffer
	if _, err := (&GraphML{World: world}).WriteTo(&out); err != nil {
		t.Fatal(err)
	}
	var file graphml.File
	if err := xml.Unmarshal(out.Bytes(), &file); err != nil {
		t.Fatal(err)
	}
	keys := map[string]string{}
	for _, key := range file.Key {
		if _, duplicate := keys[key.ID]; duplicate {
			t.Errorf("duplicate key %q", key.ID)
		}
		keys[key.ID] = key.For
	}
	checkAttrs := func(attrs []graphml.Attr, kind string) {
		for _, attr := range attrs {
			if scope := keys[attr.Key]; scope != kind && scope != "all" {
				t.Errorf("%s attribute %q has undeclared or incorrect scope %q", kind, attr.Key, scope)
			}
		}
	}
	g := file.Graphs[0]
	if len(g.Edge) != 2 {
		t.Fatalf("expected dependency and explicit link, got %d edges", len(g.Edge))
	}
	foundLabel := false
	for _, attr := range g.Edge[1].Attrs {
		if attr.Key == "label" && strings.Contains(string(attr.Value), "sends &lt;requests&gt; &amp; events") {
			foundLabel = true
		}
	}
	if !foundLabel {
		t.Error("explicit relationship label was not escaped and retained")
	}
	if len(g.Node) != 2 || g.Node[0].ID == g.Node[1].ID {
		t.Fatalf("expected two distinct nodes: %+v", g.Node)
	}
	for _, node := range g.Node {
		checkAttrs(node.Attrs, "node")
	}
	for _, edge := range g.Edge {
		checkAttrs(edge.Attrs, "edge")
		if edge.Source != g.Node[0].ID || edge.Target != g.Node[1].ID {
			t.Errorf("edge does not reference original nodes: %+v", edge)
		}
	}
	if !strings.Contains(out.String(), "xsi:schemaLocation=") {
		t.Error("missing correctly capitalized schemaLocation attribute")
	}
}

func TestELKTrimPreservesIdentity(t *testing.T) {
	a := component("example/p", "A")
	b := component("p", "A")
	a.Add(arch.NewLink("field", b))
	var out bytes.Buffer
	world := &arch.World{Components: []*arch.Component{a, b}}
	if _, err := (&ELK{World: world, Options: Options{TrimPrefix: "example/"}}).WriteTo(&out); err != nil {
		t.Fatal(err)
	}
	var nodes []string
	var source, target string
	for _, line := range strings.Split(out.String(), "\n") {
		if strings.HasPrefix(line, "node ") {
			nodes = append(nodes, strings.TrimPrefix(line, "node "))
		}
		if strings.HasPrefix(line, "edge ") {
			if _, err := fmt.Sscanf(line, "edge %s -> %s", &source, &target); err != nil {
				t.Fatal(err)
			}
		}
	}
	if len(nodes) != 2 || nodes[0] == nodes[1] || source != nodes[0] || target != nodes[1] {
		t.Fatalf("trim merged nodes or changed edge endpoints:\n%s", &out)
	}
}

func TestSanitizedNodeIdentity(t *testing.T) {
	seen := map[string]string{}
	for _, name := range []string{"a-b.Node", "a_b.Node", "a/b.Node", "a.b.Node", "a_2db.Node", "1/a.Node", "é.Node"} {
		id := sanitize(name)
		if previous, ok := seen[id]; ok {
			t.Errorf("%q and %q share ID %q", previous, name, id)
		}
		seen[id] = name
		if id[0] >= '0' && id[0] <= '9' {
			t.Errorf("ID starts with digit: %q", id)
		}
	}
	world := &arch.World{Components: []*arch.Component{component("a-b", "Node"), component("a_b", "Node")}}
	for name, writer := range map[string]io.WriterTo{"dot": &Dot{World: world}, "elk": &ELK{World: world}} {
		t.Run(name, func(t *testing.T) {
			var out bytes.Buffer
			if _, err := writer.WriteTo(&out); err != nil {
				t.Fatal(err)
			}
			for _, c := range world.Components {
				if !strings.Contains(out.String(), sanitize(c.Name())) {
					t.Errorf("missing distinct node %q:\n%s", c.Name(), &out)
				}
			}
		})
	}
}

var errWriter = errors.New("writer failed")

type limitedWriter struct {
	remaining int
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	if len(p) > w.remaining {
		n := w.remaining
		w.remaining = 0
		return n, errWriter
	}
	w.remaining -= len(p)
	return len(p), nil
}

func TestWriteToCountsAndErrors(t *testing.T) {
	world := &arch.World{Components: []*arch.Component{component("example", "Node")}}
	for name, writer := range map[string]io.WriterTo{
		"dot": &Dot{World: world}, "dot-basic": &DotBasic{World: world},
		"graphml": &GraphML{World: world}, "elk": &ELK{World: world},
	} {
		t.Run(name, func(t *testing.T) {
			var out bytes.Buffer
			n, err := writer.WriteTo(&out)
			if err != nil || n != int64(out.Len()) {
				t.Fatalf("WriteTo = (%d, %v), wrote %d bytes", n, err, out.Len())
			}
			for _, limit := range []int{0, 3, out.Len() - 1} {
				t.Run(fmt.Sprint(limit), func(t *testing.T) {
					n, err := writer.WriteTo(&limitedWriter{remaining: limit})
					if !errors.Is(err, errWriter) || n != int64(limit) {
						t.Errorf("WriteTo = (%d, %v), want (%d, writer error)", n, err, limit)
					}
				})
			}
		})
	}
}
