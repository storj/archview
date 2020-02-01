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
	write := func(format string, args ...interface{}) bool {
		if err != nil {
			return false
		}
		var wrote int
		wrote, err = fmt.Fprintf(w, format, args...)
		n += int64(wrote)
		return err == nil
	}

	write("graph G {\n")
	defer write("}\n")

	for _, source := range ctx.World.Components {
		for _, link := range source.Links {
			write("\t%q -> %q;\n", strings.TrimPrefix(source.Name(), ctx.TrimPrefix), strings.TrimPrefix(link.Target.Name(), ctx.TrimPrefix))
		}
	}
	return n, err
}
