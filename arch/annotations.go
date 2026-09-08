package arch

import (
	"go/ast"
	"strings"
	"unicode"
)

// ExtractAnnotation extracts architecture annotations from a type definition.
func ExtractAnnotation(gen *ast.GenDecl, ts *ast.TypeSpec) (comment, tag string, ok bool) {
	for _, doc := range []*ast.CommentGroup{ts.Doc, gen.Doc} {
		if doc == nil {
			continue
		}
		for i, c := range doc.List {
			kind, _ := annotationDirective(c.Text)
			if kind == "link" || kind == "resource" {
				continue
			}
			if after, ok0 := strings.CutPrefix(c.Text, "// architecture:"); ok0 {
				tag := after

				var group ast.CommentGroup
				for j, other := range doc.List {
					kind, _ := annotationDirective(other.Text)
					if j != i && kind != "link" && kind != "resource" {
						group.List = append(group.List, other)
					}
				}

				return group.Text(), strings.TrimSpace(tag), true
			}
		}
	}

	return "", "", false
}

func annotationDirective(text string) (kind, body string) {
	rest, ok := strings.CutPrefix(text, "// architecture:")
	if !ok {
		return "", ""
	}
	i := strings.IndexFunc(rest, unicode.IsSpace)
	if i < 0 {
		return rest, ""
	}
	return rest[:i], strings.TrimSpace(rest[i:])
}
