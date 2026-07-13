package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrintSimpleDiffShowsChangedAddedAndRemovedLines(t *testing.T) {
	out := captureStdout(t, func() {
		printSimpleDiff("same\nold\nremoved", "same\nnew\n")
	})

	for _, want := range []string{"-old", "+new", "-removed"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected diff output to contain %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "same") {
		t.Fatalf("unchanged lines should not be printed:\n%s", out)
	}
}

func TestOSReadFileDelegatesToFilesystem(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}

	data, err := osReadFile(path)
	if err != nil {
		t.Fatalf("read input: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("unexpected file content %q", data)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	os.Stdout = writer
	defer func() {
		os.Stdout = original
	}()

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	return buf.String()
}
