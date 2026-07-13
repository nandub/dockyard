package runner

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestDockerComposeRunnerBuildsComposeCommands(t *testing.T) {
	fake := installFakeDocker(t, 0)
	workDir := t.TempDir()
	r := DockerComposeRunner{
		WorkDir: workDir,
		Project: "test-project",
		Env:     []string{"DOCKYARD_TEST_ENV=runner-env"},
	}

	if err := r.Down(context.Background(), "compose.rendered.yaml", true); err != nil {
		t.Fatalf("down: %v", err)
	}

	log := fake.readLog(t)
	if !strings.Contains(log, "cwd="+filepath.Clean(workDir)) {
		t.Fatalf("expected command cwd %q in log:\n%s", filepath.Clean(workDir), log)
	}
	if !strings.Contains(log, "args=compose -p test-project -f compose.rendered.yaml down --volumes") {
		t.Fatalf("expected compose down arguments in log:\n%s", log)
	}
	if !strings.Contains(log, "env=runner-env") {
		t.Fatalf("expected runner env in log:\n%s", log)
	}
}

func TestDockerComposeRunnerPSIncludesAllFlag(t *testing.T) {
	fake := installFakeDocker(t, 0)
	r := DockerComposeRunner{
		WorkDir: t.TempDir(),
		Project: "test-project",
	}

	if err := r.PS(context.Background(), "compose.yaml", true); err != nil {
		t.Fatalf("ps: %v", err)
	}

	log := fake.readLog(t)
	if !strings.Contains(log, "args=compose -p test-project -f compose.yaml ps --all") {
		t.Fatalf("expected compose ps --all arguments in log:\n%s", log)
	}
}

func TestDockerComposeRunnerReturnsConciseError(t *testing.T) {
	installFakeDocker(t, 17)
	r := DockerComposeRunner{
		WorkDir: t.TempDir(),
		Project: "test-project",
	}

	err := r.Up(context.Background(), "compose.yaml")
	if err == nil {
		t.Fatal("expected docker failure")
	}
	if err.Error() != "docker command failed" {
		t.Fatalf("expected concise docker error, got %q", err.Error())
	}
}

func TestDockerPrerequisiteChecksReturnConciseErrors(t *testing.T) {
	installFakeDocker(t, 17)
	ctx := context.Background()

	checks := map[string]func(context.Context) error{
		"DockerVersion":    DockerVersion,
		"ComposeAvailable": ComposeAvailable,
		"DaemonReachable":  DaemonReachable,
	}
	expected := map[string]string{
		"DockerVersion":    "docker CLI not usable",
		"ComposeAvailable": "docker compose not available",
		"DaemonReachable":  "docker daemon not reachable",
	}

	for name, check := range checks {
		t.Run(name, func(t *testing.T) {
			err := check(ctx)
			if err == nil {
				t.Fatal("expected docker prerequisite failure")
			}
			if err.Error() != expected[name] {
				t.Fatalf("expected %q, got %q", expected[name], err.Error())
			}
		})
	}
}

type fakeDocker struct {
	logPath string
}

func installFakeDocker(t *testing.T, exitCode int) fakeDocker {
	t.Helper()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "docker.log")
	writeFakeDocker(t, dir)

	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("DOCKYARD_RUNNER_LOG", logPath)
	t.Setenv("DOCKYARD_FAKE_DOCKER_EXIT", strconv.Itoa(exitCode))

	return fakeDocker{logPath: logPath}
}

func (f fakeDocker) readLog(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(f.logPath)
	if err != nil {
		t.Fatalf("read fake docker log: %v", err)
	}
	return string(data)
}

func writeFakeDocker(t *testing.T, dir string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		path := filepath.Join(dir, "docker.bat")
		script := `@echo off
>> "%DOCKYARD_RUNNER_LOG%" echo cwd=%CD%
>> "%DOCKYARD_RUNNER_LOG%" echo args=%*
>> "%DOCKYARD_RUNNER_LOG%" echo env=%DOCKYARD_TEST_ENV%
exit /b %DOCKYARD_FAKE_DOCKER_EXIT%
`
		if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
			t.Fatalf("write fake docker: %v", err)
		}
		return
	}

	path := filepath.Join(dir, "docker")
	script := `#!/bin/sh
{
  printf 'cwd=%s\n' "$PWD"
  printf 'args=%s\n' "$*"
  printf 'env=%s\n' "$DOCKYARD_TEST_ENV"
} >> "$DOCKYARD_RUNNER_LOG"
exit "$DOCKYARD_FAKE_DOCKER_EXIT"
`
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
		t.Fatalf("write fake docker: %v", err)
	}
}
