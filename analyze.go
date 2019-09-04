package archview

import (
	"fmt"
	"go/ast"
	"go/types"
	"os"
	"strings"

	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"
)

func Analyze(pkgs ...*packages.Package) *World {
	world := NewWorld()

	packages.Visit(pkgs, func(pkg *packages.Package) bool {
		inspect := inspector.New(pkg.Syntax)
		inspect.Preorder([]ast.Node{
			(*ast.TypeSpec)(nil),
		}, func(n ast.Node) {
			ts := n.(*ast.TypeSpec)

			tag, ok := ExtractTypeTag(ts)
			if !ok {
				return
			}

			obj := pkg.TypesInfo.Defs[ts.Name]
			world.Add(&Node{
				Obj:   obj,
				Type:  obj.Type(),
				Class: tag,
			})
		})
		return true
	}, nil)

	for _, node := range world.List {
		node.Deps = findDeps("", world, node, node.Type)
	}

	return world
}

func findDeps(path string, world *World, node *Node, typ types.Type) (deps []*Dep) {
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
				deps = append(deps, findDeps(path+"."+field.Name(), world, node, f)...)
			case *types.Struct:
				deps = append(deps, findDeps(path+"."+field.Name(), world, node, f)...)
			default:
				fmt.Fprintf(os.Stderr, "unhandled method %q type %T\n", path, f)
			}
		}

	default:
		fmt.Fprintf(os.Stderr, "unhandled type %T\n", t)
	}
	return deps
}
