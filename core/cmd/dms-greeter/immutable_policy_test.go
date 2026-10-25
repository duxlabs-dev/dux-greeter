package main

import (
	"encoding/json"
	"testing"
)

// Every greeter mutation writes under /etc and /var, which image-based
// systems keep writable, so the default policy blocks nothing; images that
// want a block ship their own policy file.
func TestDefaultImmutablePolicyBlocksNothing(t *testing.T) {
	var policyFile cliPolicyFile
	if err := json.Unmarshal(defaultCLIPolicyJSON, &policyFile); err != nil {
		t.Fatalf("failed to parse embedded CLI policy: %v", err)
	}
	if policyFile.BlockedCommands == nil {
		t.Fatal("embedded CLI policy must declare blocked_commands so a packaged file can override it")
	}

	blocked := normalizeBlockedCommands(*policyFile.BlockedCommands)
	for _, cmd := range []string{"greeter install", "dms-greeter install", "greeter enable", "greeter uninstall", "greeter sync"} {
		if commandBlockedByPolicy(cmd, blocked) {
			t.Fatalf("%q must stay available on image-based systems", cmd)
		}
	}
}

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
