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
