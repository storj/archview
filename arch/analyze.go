// Package arch implements gathering component information and links.
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
				continue
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
				if _, ok := obj.Type().Underlying().(*types.Interface); !ok {
					continue
				}
				linkFrom := a.world.componentFor(obj.Type())
				if linkFrom == nil {
					continue
				}

				target := pkg.TypesInfo.TypeOf(vs.Values[i])
				if target == nil {
					continue
				}

				linkTo := a.world.componentFor(target)
				if linkTo == nil {
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
		for method := range t.Methods() {
			switch m := method.Type().(type) {
			case *types.Signature:
				result := m.Results()
				for at := range result.Variables() {

					link := a.world.componentFor(at.Type())
					if link != nil {
						source.Add(NewLink(path+"."+method.Name(), link))
					}
				}
			default:
				fmt.Fprintf(os.Stderr, "unhandled interface method type %T\n", m)
			}
		}

	case *types.Struct:
		for field := range t.Fields() {
			fieldType := tryDeref(field.Type())

			link := a.world.componentFor(fieldType)
			if link != nil {
				source.Add(NewLink(path+"."+field.Name(), link))
				continue
			}

			switch f := fieldType.Underlying().(type) {
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

	case *types.Pointer:
		a.includeLinks(source, path, t.Elem(), visiting)
	case *types.Array, *types.Signature, *types.Slice, *types.Basic, *types.Chan, *types.Map:
		// ignore basic compound types
	default:
		fmt.Fprintf(os.Stderr, "unhandled type %T\n", t)
	}
}

func tryDeref(t types.Type) types.Type {
	t = types.Unalias(t)
	for {
		ptr, ok := t.(*types.Pointer)
		if !ok {
			return t
		}
		t = types.Unalias(ptr.Elem())
	}
}
