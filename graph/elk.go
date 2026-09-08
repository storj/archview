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
	write := func(format string, args ...any) bool {
		if err != nil {
			return false
		}
		var wrote int
		wrote, err = fmt.Fprintf(w, format, args...)
		n += int64(wrote)
		return err == nil
	}

	write("algorithm: layered\n\n")

	// Prefix trimming is for readability, but must not merge different nodes.
	ids := map[*arch.Component]string{}
	used := map[string]int{}
	for _, source := range ctx.World.Components {
		base := sanitize(strings.TrimPrefix(source.Name(), ctx.TrimPrefix))
		id := base
		if used[base] > 0 {
			// A double underscore cannot occur after the prefix in sanitize's
			// output, so the suffix cannot collide with another base ID.
			id = fmt.Sprintf("%s__%d", base, used[base])
		}
		used[base]++
		ids[source] = id
		write("node %v\n", id)
	}

	write("\n")

	for _, source := range ctx.World.Components {
		for _, link := range source.Links {
			write("edge %v -> %v\n",
				ids[source],
				ids[link.Target],
			)
		}
	}
	return n, err
}
