package main

import "testing"

func TestRejectSudoInvocation(t *testing.T) {
	if err := rejectSudoInvocation(0, "purian23"); err == nil {
		t.Fatal("root with SUDO_USER must be rejected")
	}
	if err := rejectSudoInvocation(1000, ""); err != nil {
		t.Fatalf("plain user rejected: %v", err)
	}
	if err := rejectSudoInvocation(0, ""); err != nil {
		t.Fatalf("real root rejected: %v", err)
	}
}
