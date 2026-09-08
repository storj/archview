package arch

import (
	"strings"
	"testing"
)

func TestExplicitLinks(t *testing.T) {
	client := parsePackage(t, "example/client", `package client
// architecture:link Client -> example/server.Server : sends requests
// architecture:link Client -> example/server.Server : sends requests
// architecture:link Client -> Local
// architecture: Client
type Client struct{}
// architecture: Local
type Local[T any] struct{}
`)
	server := parsePackage(t, "example/server", `package server
// architecture: Server
type Server struct{}
// architecture:link Server -> example/client.Client : sends events
`)
	world, err := Analyze(client, server)
	if err != nil {
		t.Fatal(err)
	}
	if len(world.Components) != 3 {
		t.Fatalf("expected three components: %s", world)
	}
	links := world.Components[0].Links
	if len(links) != 2 || links[0].Target != world.Components[2] || links[0].Label != "sends requests" || !links[0].Explicit || links[1].Target != world.Components[1] {
		t.Fatalf("incorrect explicit links: %+v", links)
	}
	if links := world.Components[2].Links; len(links) != 1 || links[0].Target != world.Components[0] {
		t.Fatalf("reverse link not resolved: %+v", links)
	}
}

func TestExplicitLinkErrors(t *testing.T) {
	for _, directive := range []string{
		"A -> Missing", "Missing -> A", "A -> Unannotated", "A - A",
		"A ->", "A -> A :", "A -> A unexpected", "",
	} {
		t.Run(directive, func(t *testing.T) {
			pkg := parsePackage(t, "example", "package example\n// architecture: A\ntype A struct{}\ntype Unannotated struct{}\n// architecture:link "+directive+"\n")
			world, err := Analyze(pkg)
			if err == nil || world != nil || !strings.Contains(err.Error(), "example.go:5:") {
				t.Fatalf("expected positioned annotation error, got %v, %v", world, err)
			}
		})
	}
}
