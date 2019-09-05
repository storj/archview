package graph

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"

	"github.com/storj/archview/arch"
	"github.com/storj/archview/graph/graphml"
)

// GraphML implements .graphml encoding.
type GraphML struct {
	World *arch.World

	Options
}

// WriteTo writes .graphml encoding to w.
func (ctx *GraphML) WriteTo(w io.Writer) (n int64, err error) {
	file := graphml.NewFile()

	file.Graphs = append(file.Graphs, ctx.graph())

	file.Key = []graphml.Key{
		{For: "node", ID: "label", AttrName: "label", AttrType: "string"},
		{For: "node", ID: "shape", AttrName: "shape", AttrType: "string"},
		{For: "edge", ID: "label", AttrName: "label", AttrType: "string"},

		{For: "node", ID: "name", AttrName: "name", AttrType: "string"},
		{For: "node", ID: "url", AttrName: "url", AttrType: "string"},
		{For: "node", ID: "class", AttrName: "class", AttrType: "string"},
		{For: "node", ID: "package", AttrName: "package", AttrType: "string"},

		{For: "node", ID: "ynodelabel", YFilesType: "nodegraphics"},
		{For: "edge", ID: "yedgelabel", YFilesType: "edgegraphics"},
	}

	enc := xml.NewEncoder(w)
	enc.Indent("", "\t")
	return -1, enc.Encode(file)
}

func (ctx *GraphML) graph() *graphml.Graph {
	out := &graphml.Graph{}
	out.EdgeDefault = graphml.Directed

	for _, component := range ctx.World.Components {
		outnode := graphml.Node{}
		outnode.ID = ctx.id(component)

		addAttr(&outnode.Attrs, "label", component.Name())
		addAttr(&outnode.Attrs, "tooltip", component.Comment)

		addAttr(&outnode.Attrs, "name", component.ShortName())
		addAttr(&outnode.Attrs, "url", ctx.href(component))
		addAttr(&outnode.Attrs, "class", component.Class)
		addAttr(&outnode.Attrs, "package", component.Package())

		addYedLabelAttr(&outnode.Attrs, "ynodelabel", component.Name())

		out.Node = append(out.Node, outnode)
	}

	for _, src := range ctx.World.Components {
		for _, dep := range src.Deps {
			outedge := graphml.Edge{}
			outedge.Source = ctx.id(src)
			outedge.Target = ctx.id(dep.Dep)

			addAttr(&outedge.Attrs, "tooltip", dep.Path)
			out.Edge = append(out.Edge, outedge)
		}
	}

	return out
}

func (ctx *GraphML) id(component *arch.Component) string {
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

func (ctx *GraphML) href(component *arch.Component) string {
	return "http://godoc.org/" + component.Package() + "#" + component.ShortName()
}

func addAttr(attrs *[]graphml.Attr, key, value string) {
	if value == "" {
		return
	}
	*attrs = append(*attrs, graphml.Attr{
		Key:   key,
		Value: escapeText(value),
	})
}

func addYedLabelAttr(attrs *[]graphml.Attr, key, value string) {
	if value == "" {
		return
	}
	var buf bytes.Buffer
	buf.WriteString(`<y:ShapeNode><y:NodeLabel>`)
	if err := xml.EscapeText(&buf, []byte(value)); err != nil {
		// this shouldn't ever happen
		panic(err)
	}
	buf.WriteString(`</y:NodeLabel></y:ShapeNode>`)
	*attrs = append(*attrs, graphml.Attr{
		Key:   key,
		Value: buf.Bytes(),
	})
}

func escapeText(s string) []byte {
	if s == "" {
		return []byte{}
	}

	var buf bytes.Buffer
	if err := xml.EscapeText(&buf, []byte(s)); err != nil {
		// this shouldn't ever happen
		panic(err)
	}
	return buf.Bytes()
}
