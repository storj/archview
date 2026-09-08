package edit

import (
	"go/token"
	"go/types"
	"testing"

	"github.com/storj/archview/arch"
)

func TestKeepComponents(t *testing.T) {
	for _, subset := range []bool{false, true} {
		t.Run(map[bool]string{false: "separate slice", true: "shared slice"}[subset], func(t *testing.T) {
			world := arch.NewWorld()
			for _, name := range []string{"A", "B", "C"} {
				obj := types.NewTypeName(token.NoPos, nil, name, nil)
				world.Add(&arch.Component{Obj: obj, Type: types.NewNamed(obj, types.NewStruct(nil, nil), nil)})
			}
			a, b, c := world.Components[0], world.Components[1], world.Components[2]
			b.Add(arch.NewLink("a", a))
			b.Add(arch.NewLink("c", c))
			keep := []*arch.Component{b, c}
			if subset {
				keep = world.Components[1:]
			}
			KeepComponents(world, keep)
			if len(world.Components) != 2 || world.Components[0] != b || world.Components[1] != c {
				t.Fatalf("incorrect retained components: %v", world.Components)
			}
			if len(world.ByType) != 2 || world.ByType[b.Type] != b || world.ByType[c.Type] != c {
				t.Errorf("type index does not match retained components: %v", world.ByType)
			}
			if len(b.Links) != 1 || b.Links[0].Target != c {
				t.Errorf("incorrect retained links: %v", b.Links)
			}
			KeepComponents(world, nil)
			if len(world.Components) != 0 || len(world.ByType) != 0 {
				t.Error("keeping no components must empty both component collections")
			}
		})
	}
}
