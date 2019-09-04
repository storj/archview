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

	NoColor bool
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
	defer write("}\n")

	for _, component := range dot.World.Components {
		write("\t%s [%v];\n", dot.id(component),
			strings.Join([]string{
				dot.label(component), dot.href(component), dot.color(component),
			}, ","))
	}

	write("\n")

	for _, source := range dot.World.Components {
		for _, dep := range source.Deps {
			write("\t%s -> %s [%v];\n", dot.id(source), dot.id(dep.Dep),
				strings.Join([]string{
					dot.color(dep.Dep),
					dot.edgetooltip(source, dep),
				}, ","))
		}
		if len(source.Deps) > 0 {
			write("\n")
		}
	}
	return n, err
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
