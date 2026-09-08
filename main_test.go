package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBasic(t *testing.T) {
	tempdir, err := os.MkdirTemp("", "archview")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempdir)

	archviewexe := filepath.Join(tempdir, "archview.exe")

	install := exec.Command("go", "build", "-o", archviewexe, "-race", ".")
	result, err := install.CombinedOutput()
	if err != nil {
		t.Log(string(result))
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name    string
		args    []string
		message string
	}{
		{"invalid format", []string{"-format", "invalid", "./..."}, "unknown format"},
		{"missing package", []string{"./missing"}, "missing"},
		{"no matches", []string{"./...nomatch"}, "no packages matched"},
		{"unknown root", []string{"-root", "typo", "./..."}, "unknown roots"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			outfile := filepath.Join(t.TempDir(), "existing.dot")
			const original = "existing output\n"
			if err := os.WriteFile(outfile, []byte(original), 0600); err != nil {
				t.Fatal(err)
			}
			args := append([]string{"-out", outfile}, tc.args...)
			run := exec.Command(archviewexe, args...)
			run.Dir = "testdata/basic"
			output, err := run.CombinedOutput()
			if err == nil {
				t.Fatalf("expected failure, output: %s", output)
			}
			if !strings.Contains(string(output), tc.message) {
				t.Errorf("missing diagnostic %q: %s", tc.message, output)
			}
			got, err := os.ReadFile(outfile)
			if err != nil || string(got) != original {
				t.Fatalf("existing output modified: %q, error: %v", got, err)
			}
		})
	}

	for _, tc := range []struct{ name, source, diagnostic string }{
		{"invalid relationship", "package example\n// architecture: Service\ntype Service struct{}\n// architecture:link Service -> Typo\n", "Typo"},
		{"local annotation", "package example\nfunc first() {\n// architecture: Local\ntype Node struct{}\n}\n", "local type \"Node\""},
	} {
		t.Run(tc.name+" preserves output", func(t *testing.T) {
			dir := t.TempDir()
			for name, body := range map[string]string{
				"go.mod":    "module example\n\ngo 1.27\n",
				"main.go":   tc.source,
				"graph.dot": "existing output",
			} {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0600); err != nil {
					t.Fatal(err)
				}
			}
			run := exec.Command(archviewexe, "-out", "graph.dot", "./...")
			run.Dir = dir
			output, err := run.CombinedOutput()
			if err == nil || !strings.Contains(string(output), "main.go:4:") || !strings.Contains(string(output), tc.diagnostic) {
				t.Fatalf("expected positioned endpoint error: %v\n%s", err, output)
			}
			got, err := os.ReadFile(filepath.Join(dir, "graph.dot"))
			if err != nil || string(got) != "existing output" {
				t.Fatalf("output changed: %q, %v", got, err)
			}
		})
	}

	t.Run("literal root names", func(t *testing.T) {
		dir := t.TempDir()
		for name, body := range map[string]string{
			"go.mod":   "module example\n\ngo 1.27\n",
			"types.go": "package example\n// architecture: Generic\ntype Pair[A, B any] struct{}\n// architecture:resource external:users,audit : Table\n// architecture:resource external:users : Table\n",
		} {
			if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
		}
		for _, tc := range []struct {
			name  string
			roots []string
			nodes string
		}{
			{"declaration", []string{"example.Pair"}, "\t\"example.Pair[A, B any]\";\n"},
			{"display", []string{"example.Pair[A, B any]"}, "\t\"example.Pair[A, B any]\";\n"},
			{"literal comma", []string{"external:users,audit"}, "\t\"external:users,audit\";\n"},
			{"list", []string{"example.Pair, external:users"}, "\t\"example.Pair[A, B any]\";\n\t\"external:users\";\n"},
			{"repeated literals", []string{"example.Pair[A, B any]", "external:users,audit"}, "\t\"example.Pair[A, B any]\";\n\t\"external:users,audit\";\n"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				args := []string{"-format", "dot-basic"}
				for _, root := range tc.roots {
					args = append(args, "-root", root)
				}
				args = append(args, "./...")
				run := exec.Command(archviewexe, args...)
				run.Dir = dir
				got, err := run.CombinedOutput()
				if err != nil {
					t.Fatalf("root selection failed: %v\n%s", err, got)
				}
				if want := "digraph G {\n" + tc.nodes + "}\n"; string(got) != want {
					t.Fatalf("got %s, want %s", got, want)
				}
			})
		}
	})

	t.Run("client server relationships", func(t *testing.T) {
		for _, format := range []string{"dot", "dot-basic", "graphml", "elk"} {
			t.Run(format, func(t *testing.T) {
				run := exec.Command(archviewexe, "-format", format, "-root", "example/network/client.Client", "./...")
				run.Dir = "testdata/network"
				output, err := run.CombinedOutput()
				if err != nil {
					t.Fatalf("failed to generate relationships: %v\n%s", err, output)
				}
				for _, name := range []string{"Client", "Server", "payments"} {
					if !strings.Contains(string(output), name) {
						t.Errorf("missing %s in %s", name, output)
					}
				}
				if strings.Contains(string(output), "metrics") {
					t.Fatalf("unrelated resource survived root filter: %s", output)
				}
				if format != "elk" && !strings.Contains(string(output), "sends requests") {
					t.Fatalf("missing relationship label: %s", output)
				}
				if format == "dot" || format == "dot-basic" {
					if dot, err := exec.LookPath("dot"); err == nil {
						render := exec.Command(dot, "-Tsvg")
						render.Stdin = strings.NewReader(string(output))
						if out, err := render.CombinedOutput(); err != nil {
							t.Fatalf("Graphviz rejected relationships: %v\n%s", err, out)
						}
					}
				}
			})
		}
		run := exec.Command(archviewexe, "-format", "dot-basic", "-root", "example/network/client.Client", "-skip-class", "Remote Service", "./...")
		run.Dir = "testdata/network"
		output, err := run.CombinedOutput()
		if err != nil || strings.Contains(string(output), "external:") || !strings.Contains(string(output), "pushes events") {
			t.Fatalf("class filter did not preserve internal client/server cycle: %v\n%s", err, output)
		}
	})

	matches, err := filepath.Glob("testdata/*")
	if err != nil {
		t.Fatal(err)
	}

	for _, match := range matches {
		t.Run(filepath.ToSlash(match), func(t *testing.T) {
			outfile := filepath.Join(tempdir, filepath.Base(match)) + ".dot"

			run := exec.Command(archviewexe, "-format", "dot-basic", "-out", outfile, "./...")
			run.Dir = match

			result, err = run.CombinedOutput()
			t.Log(string(result))
			if err != nil {
				t.Fatal(err)
			}

			got, err := os.ReadFile(outfile)
			if err != nil {
				t.Fatalf("failed to read output: %v", err)
			}
			expected, err := os.ReadFile(filepath.Join(match, "graph.dot"))
			if err != nil {
				t.Fatalf("failed to read expected: %v", err)
			}

			if diff := cmp.Diff(expected, got); diff != "" {
				t.Error(diff)
				t.Fatal("output does not match the expectation")
			}
		})
	}
}
