package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type renderFunc func(io.Writer) (int64, error)

func (f renderFunc) WriteTo(w io.Writer) (int64, error) { return f(w) }

func TestWriteOutput(t *testing.T) {
	for _, failure := range []string{"", "render", "closed file"} {
		t.Run("failure="+failure, func(t *testing.T) {
			dir := t.TempDir()
			name := filepath.Join(dir, "graph.dot")
			if err := os.WriteFile(name, []byte("original"), 0640); err != nil {
				t.Fatal(err)
			}
			original, err := os.Stat(name)
			if err != nil {
				t.Fatal(err)
			}
			broken := errors.New("render failed")
			err = writeOutput(name, renderFunc(func(w io.Writer) (int64, error) {
				n, err := io.WriteString(w, "replacement")
				if err != nil {
					return int64(n), err
				}
				switch failure {
				case "render":
					return int64(n), broken
				case "closed file":
					return int64(n), w.(io.Closer).Close()
				}
				return int64(n), nil
			}))
			want := "replacement"
			if failure != "" {
				want = "original"
				if err == nil {
					t.Fatal("expected output failure")
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if failure == "render" && !errors.Is(err, broken) {
				t.Fatalf("lost render error: %v", err)
			}
			got, err := os.ReadFile(name)
			if err != nil || string(got) != want {
				t.Fatalf("output = %q, %v; want %q", got, err, want)
			}
			info, err := os.Stat(name)
			if err != nil {
				t.Fatal(err)
			}
			if info.Mode().Perm() != original.Mode().Perm() {
				t.Error("changed destination permissions")
			}
			files, err := os.ReadDir(dir)
			if err != nil || len(files) != 1 {
				t.Fatalf("temporary file leaked: %v, %v", files, err)
			}
		})
	}
}

func TestWriteOutputNewFileAndSymlink(t *testing.T) {
	dir := t.TempDir()
	name := filepath.Join(dir, "graph.dot")
	if err := writeOutput(name, strings.NewReader("graph")); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(name); err != nil || string(got) != "graph" {
		t.Fatalf("new output = %q, %v", got, err)
	}
	if runtime.GOOS == "windows" {
		return
	} // Symlink creation can require privileges.
	link := filepath.Join(dir, "link.dot")
	if err := os.Symlink(name, link); err != nil {
		t.Fatal(err)
	}
	if err := writeOutput(link, strings.NewReader("replacement")); err == nil {
		t.Fatal("expected symlink rejection")
	}
	if got, err := os.ReadFile(name); err != nil || string(got) != "graph" {
		t.Fatalf("symlink target changed: %q, %v", got, err)
	}
}
