package arch

import (
	"fmt"
	"go/ast"
	"go/types"
	"os"

	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"
)

type analysis struct {
	pkgs  []*packages.Package
	world *World
}

// Analyze analyzes packages and extracts architecture annotations and relations.
func Analyze(pkgs ...*packages.Package) *World {
	a := &analysis{
		pkgs:  pkgs,
		world: NewWorld(),
	}
	packages.Visit(pkgs, a.visitpkg, nil)

	for _, component := range a.world.Components {
		a.includeDeps(component, "", component.Type, visiting{})
	}

	return a.world
}

func (a *analysis) visitpkg(pkg *packages.Package) bool {
	inspect := inspector.New(pkg.Syntax)
	inspect.Preorder([]ast.Node{
		(*ast.GenDecl)(nil),
	}, func(n ast.Node) {
		gen := n.(*ast.GenDecl)
		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			comment, tag, ok := ExtractAnnotation(gen, ts)
			if !ok {
				return
			}

			obj := pkg.TypesInfo.Defs[ts.Name]
			a.world.Add(&Component{
				Obj:     obj,
				Type:    obj.Type(),
				Class:   tag,
				Comment: comment,
			})
		}
	})
	return true
}

type visiting map[types.Type]struct{}

// includeDeps adds dependencies to source.
func (a *analysis) includeDeps(source *Component, path string, typ types.Type, visiting visiting) {
	if _, ok := visiting[typ]; ok {
		return
	}
	visiting[typ] = struct{}{}
	defer delete(visiting, typ)

	switch t := typ.Underlying().(type) {
	case *types.Interface:
		for i := 0; i < t.NumMethods(); i++ {
			method := t.Method(i)
			switch m := method.Type().(type) {
			case *types.Signature:
				result := m.Results()
				for i := 0; i < result.Len(); i++ {
					at := result.At(i)

					dep := a.world.ByType[at.Type().Underlying()]
					if dep != nil {
						source.Add(NewDep(path+"."+method.Name(), dep))
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

			dep := a.world.ByType[underlying]
			if dep != nil {
				source.Add(NewDep(path+"."+field.Name(), dep))
				continue
			}

			switch f := underlying.(type) {
			case *types.Pointer:
				a.includeDeps(source, path+"."+field.Name(), f, visiting)
			case *types.Struct:
				a.includeDeps(source, path+"."+field.Name(), f, visiting)
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
