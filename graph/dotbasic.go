package graph

import (
	"fmt"
	"io"
	"strings"

	"github.com/storj/archview/arch"
)

// DotBasic implements basic .dot encoding.
type DotBasic struct {
	World *arch.World

	Options
}

// WriteTo writes dot output to w.
func (ctx *DotBasic) WriteTo(w io.Writer) (n int64, err error) {
	write := func(format string, args ...any) bool {
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

	connected := map[*arch.Component]bool{}
	for _, source := range ctx.World.Components {
		for _, link := range source.Links {
			connected[source] = true
			connected[link.Target] = true
		}
	}
	for _, component := range ctx.World.Components {
		label := strings.TrimPrefix(component.Name(), ctx.TrimPrefix)
		if label != component.Name() {
			write("\t%q [label=%q];\n", component.Name(), label)
		} else if !connected[component] {
			write("\t%q;\n", component.Name())
		}
	}

	for _, source := range ctx.World.Components {
		for _, link := range source.Links {
			if link.Label != "" {
				write("\t%q -> %q [label=%q];\n", source.Name(), link.Target.Name(), link.Label)
			} else {
				write("\t%q -> %q;\n", source.Name(), link.Target.Name())
			}
		}
	}
	return n, err
}
