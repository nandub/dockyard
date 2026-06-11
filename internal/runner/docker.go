package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

type DockerComposeRunner struct {
	WorkDir string
	Project string
	Env     []string
}

func (r DockerComposeRunner) Up(ctx context.Context, composeFile string) error {
	return r.run(ctx, "compose", "-p", r.Project, "-f", composeFile, "up", "-d")
}

func (r DockerComposeRunner) Down(ctx context.Context, composeFile string, removeVolumes bool) error {
	args := []string{"compose", "-p", r.Project, "-f", composeFile, "down"}
	if removeVolumes {
		args = append(args, "--volumes")
	}
	return r.run(ctx, args...)
}

func (r DockerComposeRunner) Config(ctx context.Context, composeFile string) error {
	return r.runWithStdout(ctx, os.Stdout, "compose", "-p", r.Project, "-f", composeFile, "config")
}

func (r DockerComposeRunner) ValidateConfig(ctx context.Context, composeFile string) error {
	return r.runWithStdout(ctx, io.Discard, "compose", "-p", r.Project, "-f", composeFile, "config")
}

func (r DockerComposeRunner) PS(ctx context.Context, composeFile string, all bool) error {
	args := []string{"compose", "-p", r.Project, "-f", composeFile, "ps"}
	if all {
		args = append(args, "--all")
	}
	return r.run(ctx, args...)
}

func (r DockerComposeRunner) run(ctx context.Context, args ...string) error {
	return r.runWithStdout(ctx, os.Stdout, args...)
}

func (r DockerComposeRunner) runWithStdout(ctx context.Context, stdout io.Writer, args ...string) error {
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = filepath.Clean(r.WorkDir)
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr
	if len(r.Env) > 0 {
		cmd.Env = append(os.Environ(), r.Env...)
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker command failed")
	}
	return nil
}

func CommandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func DockerVersion(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker CLI not usable")
	}
	return nil
}

func ComposeAvailable(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "compose", "version")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker compose not available")
	}
	return nil
}

func DaemonReachable(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "info")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker daemon not reachable")
	}
	return nil
}
