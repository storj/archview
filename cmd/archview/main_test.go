package main_test

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func Test(t *testing.T) {
	install := exec.Command("go", "install", "-race", ".")
	result, err := install.CombinedOutput()
	if err != nil {
		t.Log(string(result))
		t.Fatal(err)
	}

	run := exec.Command("archview", "./...")
	run.Dir, _ = filepath.Abs("../../testdata/src/example")

	result, err = run.CombinedOutput()
	t.Log(string(result))
	if err != nil {
		t.Fatal(err)
	}
}
