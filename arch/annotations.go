package arch

import (
	"go/ast"
	"strings"
)

func ExtractAnnotation(ts *ast.TypeSpec) (tag string, ok bool) {
	if ts.Comment == nil {
		return "", false
	}

	for _, c := range ts.Comment.List {
		if strings.HasPrefix(c.Text, "// archview:") {
			tag := strings.TrimPrefix(c.Text, "// archview:")
			return strings.TrimSpace(tag), true
		}
	}

	return "", false
}
