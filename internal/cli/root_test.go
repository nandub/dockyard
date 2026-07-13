package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRootCommandRegistersExpectedTopLevelCommands(t *testing.T) {
	cmd := NewRootCommand()
	if cmd.Use != "dockyard" {
		t.Fatalf("unexpected root use: %q", cmd.Use)
	}
	if flag := cmd.PersistentFlags().Lookup("home"); flag == nil {
		t.Fatal("expected root --home persistent flag")
	}

	want := []string{
		"catalog", "compat", "config", "diff", "doctor", "env", "init", "inspect",
		"install", "install-plan", "lint", "list", "lock", "package", "policy",
		"prune", "pull", "push", "rollback", "render", "secrets", "status",
		"uninstall", "upgrade", "values", "verify", "version",
	}
	for _, name := range want {
		if _, _, err := cmd.Find([]string{name}); err != nil {
			t.Fatalf("expected root command %q to be registered: %v", name, err)
		}
	}
}

func TestVersionCommandPrintsTextAndJSON(t *testing.T) {
	text := captureStdout(t, func() {
		cmd := newVersionCommand()
		cmd.SetArgs(nil)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("execute version command: %v", err)
		}
	})
	if !strings.Contains(text, "dockyard version") || !strings.Contains(text, "commit") {
		t.Fatalf("unexpected version text output:\n%s", text)
	}

	jsonText := captureStdout(t, func() {
		cmd := newVersionCommand()
		cmd.SetArgs([]string{"--json"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("execute version --json command: %v", err)
		}
	})
	var parsed map[string]any
	if err := json.Unmarshal([]byte(jsonText), &parsed); err != nil {
		t.Fatalf("expected JSON version output, got %q: %v", jsonText, err)
	}
	if parsed["version"] == "" || parsed["go"] == "" {
		t.Fatalf("expected version metadata fields, got %#v", parsed)
	}
}

func TestPolicyCommandRegistersListAndCheckSubcommands(t *testing.T) {
	cmd := newPolicyCommand()
	for _, name := range []string{"list", "check"} {
		if _, _, err := cmd.Find([]string{name}); err != nil {
			t.Fatalf("expected policy subcommand %q: %v", name, err)
		}
	}
}

func TestPolicyListCommandPrintsTextAndJSON(t *testing.T) {
	text := captureStdout(t, func() {
		cmd := newPolicyListCommand()
		if err := cmd.Execute(); err != nil {
			t.Fatalf("execute policy list: %v", err)
		}
	})
	if !strings.Contains(text, "privileged mode is not allowed") {
		t.Fatalf("unexpected policy list text output:\n%s", text)
	}

	jsonText := captureStdout(t, func() {
		cmd := newPolicyListCommand()
		cmd.SetArgs([]string{"--json"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("execute policy list --json: %v", err)
		}
	})
	var parsed []map[string]any
	if err := json.Unmarshal([]byte(jsonText), &parsed); err != nil {
		t.Fatalf("expected JSON policy list output, got %q: %v", jsonText, err)
	}
	if len(parsed) == 0 {
		t.Fatal("expected JSON policy list entries")
	}
}
