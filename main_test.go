package main_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBasic(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "archview")
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

			got, err := ioutil.ReadFile(outfile)
			if err != nil {
				t.Fatalf("failed to read output: %v", err)
			}
			expected, err := ioutil.ReadFile(filepath.Join(match, "graph.dot"))
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
