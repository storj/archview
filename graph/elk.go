package graph

import (
	"fmt"
	"io"
	"strings"

	"github.com/storj/archview/arch"
)

// ELK implements basic ELK compatible encoding.
type ELK struct {
	World *arch.World

	Options
}

// WriteTo writes dot output to w.
func (ctx *ELK) WriteTo(w io.Writer) (n int64, err error) {
	write := func(format string, args ...interface{}) bool {
		if err != nil {
			return false
		}
		var wrote int
		wrote, err = fmt.Fprintf(w, format, args...)
		n += int64(wrote)
		return err == nil
	}

	write("algorithm: layered\n\n")

	for _, source := range ctx.World.Components {
		if ctx.Skip(source) {
			continue
		}
		write("node %v\n", sanitize(strings.TrimPrefix(source.Name(), ctx.TrimPrefix)))
	}

	write("\n")

	for _, source := range ctx.World.Components {
		if ctx.Skip(source) {
			continue
		}
		for _, link := range source.Links {
			if ctx.Skip(link.Target) {
				continue
			}

			write("edge %v -> %v\n",
				sanitize(strings.TrimPrefix(source.Name(), ctx.TrimPrefix)),
				sanitize(strings.TrimPrefix(link.Target.Name(), ctx.TrimPrefix)),
			)
		}
	}
	return n, err
}
