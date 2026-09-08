package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// writeOutput stages a complete graph beside the destination before replacing it.
func writeOutput(name string, renderer io.WriterTo) error {
	info, err := os.Lstat(name)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if info != nil && !info.Mode().IsRegular() {
		return fmt.Errorf("output %q must be a regular file", name)
	}
	file, err := os.CreateTemp(filepath.Dir(name), ".archview-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	if _, err := renderer.WriteTo(file); err != nil {
		return err
	}
	if info != nil {
		if err := file.Chmod(info.Mode().Perm()); err != nil {
			return err
		}
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), name)
}
