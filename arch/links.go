package arch

import (
	"fmt"
	"go/token"
	"strings"

	"golang.org/x/tools/go/packages"
)

type explicitLink struct {
	from, to, label string
	pkg             string
	position        token.Position
}

// addAnnotations reads standalone directives as well as type documentation.
func (a *analysis) addAnnotations(pkg *packages.Package) bool {
	for _, file := range pkg.Syntax {
		for _, group := range file.Comments {
			for _, comment := range group.List {
				kind, body := annotationDirective(comment.Text)
				if kind != "link" && kind != "resource" {
					continue
				}
				var position token.Position
				if pkg.Fset != nil {
					position = pkg.Fset.Position(comment.Pos())
				}
				fields := strings.Fields(body)
				if kind == "resource" {
					a.addResource(fields, position)
					continue
				}
				if len(fields) < 3 || fields[1] != "->" || (len(fields) > 3 && (len(fields) < 5 || fields[3] != ":")) {
					a.errors = append(a.errors, fmt.Errorf("%s: expected architecture:link Source -> Target [: label]", position))
					continue
				}
				label := ""
				if len(fields) > 3 {
					label = strings.Join(fields[4:], " ")
				}
				a.links = append(a.links, explicitLink{from: fields[0], to: fields[2], label: label, pkg: pkg.Types.Path(), position: position})
			}
		}
	}
	return true
}

func (a *analysis) addResource(fields []string, position token.Position) {
	if len(fields) < 3 || fields[1] != ":" || !strings.HasPrefix(fields[0], "external:") || fields[0] == "external:" {
		a.errors = append(a.errors, fmt.Errorf("%s: expected architecture:resource external:ID : Class", position))
		return
	}
	id, class := fields[0], strings.Join(fields[2:], " ")
	if existing := a.resources[id]; existing != nil {
		if existing.Class != class {
			a.errors = append(a.errors, fmt.Errorf("%s: conflicting classes %q and %q for resource %q", position, existing.Class, class, id))
		}
		return
	}
	resource := &Component{ExternalID: id, Class: class}
	a.resources[id] = resource
	a.world.Add(resource)
}

func (a *analysis) resolveLinks() {
	byName := map[string][]*Component{}
	for _, component := range a.world.Components {
		// Declaration names also address generic components without type parameters.
		name := component.ReferenceName()
		byName[name] = append(byName[name], component)
	}
	for _, pending := range a.links {
		resolve := func(name string) *Component {
			if !strings.Contains(name, ".") && !strings.HasPrefix(name, "external:") {
				name = pending.pkg + "." + name
			}
			matches := byName[name]
			if len(matches) != 1 {
				reason := "unknown"
				if len(matches) > 1 {
					reason = "ambiguous"
				}
				a.errors = append(a.errors, fmt.Errorf("%s: %s architecture link endpoint %q (expected an annotated type or declared resource)", pending.position, reason, name))
				return nil
			}
			return matches[0]
		}
		from, to := resolve(pending.from), resolve(pending.to)
		if from == nil || to == nil {
			continue
		}
		duplicate := false
		for _, link := range from.Links {
			if link.Explicit && link.Target == to && link.Label == pending.label {
				duplicate = true
				break
			}
		}
		if !duplicate {
			from.Add(&Link{Target: to, Label: pending.label, Explicit: true})
		}
	}
}
