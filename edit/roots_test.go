package edit

import (
	"go/token"
	"go/types"
	"strings"
	"testing"

	"github.com/storj/archview/arch"
)

func TestExternalResourceFiltering(t *testing.T) {
	world := arch.NewWorld()
	client := &arch.Component{ExternalID: "external:client", Class: "Client"}
	server := &arch.Component{ExternalID: "external:server", Class: "Server"}
	table := &arch.Component{ExternalID: "external:table", Class: "Table"}
	unused := &arch.Component{ExternalID: "external:unused", Class: "Table"}
	for _, c := range []*arch.Component{client, server, table, unused} {
		world.Add(c)
	}
	client.Add(&arch.Link{Target: server, Explicit: true})
	server.Add(&arch.Link{Target: table, Explicit: true})
	if err := KeepRoots(world, []string{"external:client"}); err != nil {
		t.Fatal(err)
	}
	if len(world.Components) != 3 {
		t.Fatalf("explicit transitive dependencies not retained: %s", world)
	}
	RemoveClasses(world, []string{"Table"})
	if len(world.Components) != 2 || len(server.Links) != 0 || len(client.Links) != 1 || len(world.ByType) != 0 {
		t.Fatalf("external filtering left inconsistent graph: %s", world)
	}
}

func TestKeepRoots(t *testing.T) {
	world := arch.NewWorld()
	for _, name := range []string{"A", "B", "C"} {
		obj := types.NewTypeName(token.NoPos, nil, name, nil)
		world.Add(&arch.Component{Obj: obj, Type: types.NewNamed(obj, types.NewStruct(nil, nil), nil)})
	}
	a, b := world.Components[0], world.Components[1]
	a.Add(arch.NewLink("b", b))
	b.Add(arch.NewLink("a", a))
	err := KeepRoots(world, []string{"A", "typo"})
	if err == nil || !strings.Contains(err.Error(), "typo") || !strings.Contains(err.Error(), "A, B, C") {
		t.Fatalf("expected unknown root and available names, got %v", err)
	}
	if len(world.Components) != 3 || len(world.ByType) != 3 || len(a.Links) != 1 || len(b.Links) != 1 {
		t.Fatal("invalid roots changed world")
	}
	if err := KeepRoots(world, []string{"A"}); err != nil {
		t.Fatal(err)
	}
	if len(world.Components) != 2 || world.Components[0] != a || world.Components[1] != b {
		t.Fatalf("root and cyclic dependency not retained: %v", world.Components)
	}
}
