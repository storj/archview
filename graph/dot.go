package graph

import (
	"crypto/sha256"
	"fmt"
	"io"
	"strings"

	"github.com/storj/archview/arch"
)

type Dot struct {
	World *arch.World

	GroupByClass bool
	NoColor      bool
}

func (dot *Dot) WriteTo(w io.Writer) (n int64, err error) {
	write := func(format string, args ...interface{}) bool {
		if err != nil {
			return false
		}
		var wrote int
		wrote, err = fmt.Fprintf(w, format, args...)
		n += int64(wrote)
		return err == nil
	}

	write("digraph G {\n")

	if dot.NoColor {
		write("\tnode [shape=record target=\"_graphviz\"];\n")
		write("\tedge [];\n")
	} else {
		write("\tnode [penwidth=2 shape=record target=\"_graphviz\"];\n")
		write("\tedge [penwidth=2];\n")
	}
	write("\tcompound=true;\n")

	write("\trankdir=LR;\n")
	write("\tnewrank=true;\n")

	write("\n")
	defer write("}\n")

	if dot.GroupByClass {
		byClass := map[string][]*arch.Component{}
		for _, component := range dot.World.Components {
			byClass[component.Class] = append(byClass[component.Class], component)
		}

		for class, components := range byClass {
			write("\tsubgraph cluster_%v {\n", class)
			write("\t\tlabel=%q;\n\n", class)
			write("\t\tfontsize=10;\n\n")
			for _, component := range components {
				write("\t\t%s %v;\n", dot.id(component),
					attrs(
						dot.label(component),
						dot.href(component),
						dot.color(component),
						dot.nodetooltip(component),
					))
			}
			write("\t}\n")
		}
	} else {
		for _, component := range dot.World.Components {
			write("\t%s %v;\n", dot.id(component),
				attrs(
					dot.label(component),
					dot.href(component),
					dot.color(component),
				))
		}
	}

	write("\n")

	for _, source := range dot.World.Components {
		for _, dep := range source.Deps {
			write("\t%s -> %s %v;\n", dot.id(source), dot.id(dep.Dep),
				attrs(
					dot.color(dep.Dep),
					dot.edgetooltip(source, dep),
				))
		}
		if len(source.Deps) > 0 {
			write("\n")
		}
	}
	return n, err
}

func attrs(list ...string) string {
	xs := list[:0]
	for _, x := range list {
		if x != "" {
			xs = append(xs, x)
		}
	}
	if len(xs) == 0 {
		return ""
	}
	return "[" + strings.Join(xs, ",") + "]"
}

func (dot *Dot) id(component *arch.Component) string {
	return strings.Map(func(r rune) rune {
		switch {
		case 'a' <= r && r <= 'z':
			return r
		case 'A' <= r && r <= 'Z':
			return r
		case '0' <= r && r <= '9':
			return r
		default:
			return '_'
		}
	}, component.Name())
}

func (dot *Dot) label(component *arch.Component) string {
	return fmt.Sprintf("label=%q", component.Name())
}

func (dot *Dot) nodetooltip(component *arch.Component) string {
	return fmt.Sprintf("tooltip=%q", component.Comment)
}

func (dot *Dot) edgetooltip(source *arch.Component, dep *arch.Dep) string {
	return fmt.Sprintf("tooltip=%q", dep.Path)
}

func (dot *Dot) href(component *arch.Component) string {
	return fmt.Sprintf("href=%q", "http://godoc.org/"+component.PkgPath()+"#"+component.ShortName())
}

func (dot *Dot) color(component *arch.Component) string {
	if dot.NoColor {
		return ""
	}

	hash := sha256.Sum256([]byte(component.Name()))
	hue := float64(uint(hash[0])<<8|uint(hash[1])) / 0xFFFF
	return "color=" + hslahex(hue, 0.9, 0.3, 0.7)
}
