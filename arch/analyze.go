package arch

import (
	"fmt"
	"go/ast"
	"go/types"
	"os"

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

			tag, ok := ExtractAnnotation(ts)
			if !ok {
				return
			}

			obj := pkg.TypesInfo.Defs[ts.Name]
			world.Add(&Component{
				Obj:   obj,
				Type:  obj.Type(),
				Class: tag,
			})
		})
		return true
	}, nil)

	for _, component := range world.Components {
		includeDeps("", world, component, component.Type)
	}

	return world
}

func includeDeps(path string, world *World, source *Component, typ types.Type) {
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
						source.Add(path+"."+method.Name(), dep)
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
				source.Add(path+"."+field.Name(), dep)
				continue
			}

			switch f := underlying.(type) {
			case *types.Pointer:
				includeDeps(path+"."+field.Name(), world, source, f)
			case *types.Struct:
				includeDeps(path+"."+field.Name(), world, source, f)
			default:
				fmt.Fprintf(os.Stderr, "unhandled method %q type %T\n", path, f)
			}
		}

	default:
		fmt.Fprintf(os.Stderr, "unhandled type %T\n", t)
	}
}

func tryDeref(t types.Type) types.Type {
	if ptr, ok := t.(*types.Pointer); ok {
		return ptr.Elem().Underlying()
	}
	return t.Underlying()
}
