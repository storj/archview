package arch

import (
	"fmt"
	"go/ast"
	"strings"
)

func ExtractAnnotation(gen *ast.GenDecl, ts *ast.TypeSpec) (comment, tag string, ok bool) {
	fmt.Printf("===\n")
	fmt.Printf("%v\n", gen.Doc.Text())

	fmt.Printf("%v\n", ts.Name)
	fmt.Printf("%v\n", ts.Doc.Text())
	fmt.Printf("%v\n", ts.Comment.Text())

	if gen.Doc == nil {
		return "", "", false
	}

	for i, c := range gen.Doc.List {
		if strings.HasPrefix(c.Text, "// architecture:") {
			tag := strings.TrimPrefix(c.Text, "// architecture:")

			var group ast.CommentGroup
			group.List = append(group.List, gen.Doc.List[:i]...)
			group.List = append(group.List, gen.Doc.List[i+1:]...)

			return group.Text(), strings.TrimSpace(tag), true
		}
	}

	return "", "", false
}
