// Package arch implements gathering component information and links.
package arch

import (
	"go/ast"
	"go/types"

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
			a.includeLinks(source, path+"."+method.Name(), method.Type(), visiting)
		}
	case *types.Signature:
		for _, tuple := range []*types.Tuple{t.Params(), t.Results()} {
			for variable := range tuple.Variables() {
				a.includeDependency(source, path, variable.Type(), visiting)
			}
		}
	case *types.Struct:
		for field := range t.Fields() {
			a.includeDependency(source, path+"."+field.Name(), field.Type(), visiting)
		}
	case *types.Pointer:
		a.includeDependency(source, path, t.Elem(), visiting)
	case *types.Array:
		a.includeDependency(source, path, t.Elem(), visiting)
	case *types.Slice:
		a.includeDependency(source, path, t.Elem(), visiting)
	case *types.Chan:
		a.includeDependency(source, path, t.Elem(), visiting)
	case *types.Map:
		a.includeDependency(source, path, t.Key(), visiting)
		a.includeDependency(source, path, t.Elem(), visiting)
	}
}

func (a *analysis) includeDependency(source *Component, path string, typ types.Type, visiting visiting) {
	if target := a.world.componentFor(typ); target != nil {
		link := NewLink(path, target)
		for _, existing := range source.Links {
			if !existing.Implementation && existing.Path == link.Path && existing.Target == target {
				return
			}
		}
		source.Add(link)
		return
	}
	a.includeLinks(source, path, typ, visiting)
}
