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

	packages.Visit(pkgs, a.addComponents, nil)

	for _, component := range a.world.Components {
		a.includeLinks(component, "", component.Type, visiting{})
	}

	packages.Visit(pkgs, a.addImpls, nil)

	return a.world
}

func (a *analysis) addComponents(pkg *packages.Package) bool {
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

func (a *analysis) addImpls(pkg *packages.Package) bool {
	inspect := inspector.New(pkg.Syntax)
	inspect.Preorder([]ast.Node{
		(*ast.GenDecl)(nil),
	}, func(n ast.Node) {
		gen := n.(*ast.GenDecl)
		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			// ignore func calls
			if len(vs.Names) != len(vs.Values) {
				continue
			}

			for i, source := range vs.Names {
				obj := pkg.TypesInfo.ObjectOf(source)
				linkFrom, isComponent := a.world.ByType[tryDeref(obj.Type().Underlying())]
				if !isComponent {
					continue
				}

				target := pkg.TypesInfo.TypeOf(vs.Values[i])
				if target == nil {
					continue
				}

				linkTo, isComponent := a.world.ByType[tryDeref(target.Underlying())]
				if !isComponent {
					continue
				}

				linkFrom.Add(NewImplLink(linkTo))
			}
		}
	})
	return true
}

type visiting map[types.Type]struct{}

// includeLinks adds dependencies to source.
func (a *analysis) includeLinks(source *Component, path string, typ types.Type, visiting visiting) {
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

					link := a.world.ByType[at.Type().Underlying()]
					if link != nil {
						source.Add(NewLink(path+"."+method.Name(), link))
					}
				}
			default:
				fmt.Fprintf(os.Stderr, "unhandled interface method type %T\n", m)
			}
		}

	case *types.Struct:
		for i := 0; i < t.NumFields(); i++ {
			field := t.Field(i)
			underlying := tryDeref(field.Type().Underlying())

			link := a.world.ByType[underlying]
			if link != nil {
				source.Add(NewLink(path+"."+field.Name(), link))
				continue
			}

			switch f := underlying.(type) {
			case *types.Pointer:
				a.includeLinks(source, path+"."+field.Name(), f, visiting)
			case *types.Struct:
				a.includeLinks(source, path+"."+field.Name(), f, visiting)
			case *types.Interface:
				a.includeLinks(source, path+"."+field.Name(), f, visiting)

			case *types.Array, *types.Signature, *types.Slice, *types.Basic, *types.Chan, *types.Map:
				// ignore basic compound types
			default:
				fmt.Fprintf(os.Stderr, "unhandled struct field %q type %T\n", path, f)
			}
		}

	case *types.Array, *types.Signature, *types.Slice, *types.Basic, *types.Chan, *types.Map:
		// ignore basic compound types
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
