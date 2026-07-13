package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
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

func TestRepresentativeCommandArgumentContracts(t *testing.T) {
	tests := []struct {
		name string
		cmd  func() *cobra.Command
		args []string
	}{
		{name: "init requires directory", cmd: func() *cobra.Command { return newInitCommand() }, args: nil},
		{name: "lint requires package dir", cmd: func() *cobra.Command { return newLintCommand() }, args: nil},
		{name: "render requires package dir", cmd: func() *cobra.Command { return newRenderCommand() }, args: nil},
		{name: "pull requires OCI reference", cmd: func() *cobra.Command { return newPullCommand() }, args: nil},
		{name: "push requires archive and ref", cmd: func() *cobra.Command { return newPushCommand() }, args: []string{"archive.tgz"}},
		{name: "install-plan requires release and source", cmd: func() *cobra.Command { return newInstallPlanCommand(&globalOptions{}) }, args: []string{"release"}},
		{name: "rollback requires release and revision", cmd: func() *cobra.Command { return newRollbackCommand(&globalOptions{}) }, args: []string{"release"}},
		{name: "status requires release", cmd: func() *cobra.Command { return newStatusCommand(&globalOptions{}) }, args: nil},
		{name: "uninstall requires release", cmd: func() *cobra.Command { return newUninstallCommand(&globalOptions{}) }, args: nil},
		{name: "verify requires archive", cmd: func() *cobra.Command { return newVerifyCommand() }, args: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmd()
			if cmd.Args == nil {
				t.Fatalf("%s has no argument validator", tt.name)
			}
			if err := cmd.Args(cmd, tt.args); err == nil {
				t.Fatalf("expected argument validation error for args %#v", tt.args)
			}
		})
	}
}
