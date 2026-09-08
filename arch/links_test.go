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
	if world.Components[0].Class != "Client" {
		t.Fatalf("link directive was mistaken for the component class: %q", world.Components[0].Class)
	}
	if strings.Contains(world.Components[0].Comment, "architecture:") {
		t.Fatalf("annotation leaked into documentation: %q", world.Components[0].Comment)
	}
	links := world.Components[0].Links
	if len(links) != 2 || links[0].Target != world.Components[2] || links[0].Label != "sends requests" || !links[0].Explicit || links[1].Target != world.Components[1] {
		t.Fatalf("incorrect explicit links: %+v", links)
	}
	if links := world.Components[2].Links; len(links) != 1 || links[0].Target != world.Components[0] {
		t.Fatalf("reverse link not resolved: %+v", links)
	}
}

func TestExternalResources(t *testing.T) {
	db := parsePackage(t, "example/db", `package db
// architecture: Database
type DB struct{}
// architecture:link DB -> external:users : writes
// architecture:link external:users -> external:audit : replicates
// architecture:resource external:users : Database Table
`)
	other := parsePackage(t, "example/other", `package other
// architecture:resource external:users : Database Table
// architecture:resource external:audit : Remote Service
`)
	world, err := Analyze(db, other)
	if err != nil {
		t.Fatal(err)
	}
	if len(world.Components) != 3 || len(world.ByType) != 1 {
		t.Fatalf("unexpected resource/index counts: %s", world)
	}
	users, audit := world.Components[1], world.Components[2]
	if users.Name() != "external:users" || users.ShortName() != "users" || users.Class != "Database Table" || users.Package() != "" || users.Type != nil || users.Obj != nil {
		t.Fatalf("invalid external resource: %+v", users)
	}
	if len(world.Components[0].Links) != 1 || world.Components[0].Links[0].Target != users || len(users.Links) != 1 || users.Links[0].Target != audit {
		t.Fatalf("external relationships not resolved: %s", world)
	}
}

func TestResourceErrors(t *testing.T) {
	for _, directive := range []string{
		"// architecture:resource users : Table",
		"// architecture:resource external: : Table",
		"// architecture:resource external:users :",
		"// architecture:resource external:users Table",
		"// architecture:resource external:users : Table\n// architecture:resource external:users : Service",
		"// architecture:link external:missing -> external:unknown",
	} {
		t.Run(directive, func(t *testing.T) {
			world, err := Analyze(parsePackage(t, "example", "package example\n"+directive+"\n"))
			if err == nil || world != nil || !strings.Contains(err.Error(), "example.go:") {
				t.Fatalf("expected positioned resource error, got %v, %v", world, err)
			}
		})
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
