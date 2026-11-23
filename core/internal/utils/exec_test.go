package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLookCommandFindsSbinOffPath(t *testing.T) {
	sbin := t.TempDir()
	tool := filepath.Join(sbin, "semanage")
	if err := os.WriteFile(tool, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	saved := systemSbinDirs
	systemSbinDirs = []string{sbin}
	t.Cleanup(func() { systemSbinDirs = saved })
	t.Setenv("PATH", t.TempDir())

	got, ok := LookCommand("semanage")
	if !ok || got != tool {
		t.Fatalf("LookCommand(semanage) = %q, %v; want %q, true", got, ok, tool)
	}
	if !CommandExists("semanage") {
		t.Fatal("CommandExists(semanage) = false, want true")
	}
	if _, ok := LookCommand("not-a-real-tool"); ok {
		t.Fatal("LookCommand found a missing tool")
	}
}
