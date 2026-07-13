package version

import (
	"runtime"
	"strings"
	"testing"
)

func TestGetReturnsBuildAndRuntimeMetadata(t *testing.T) {
	info := Get()
	if info.Version != Version || info.Commit != Commit || info.Date != Date {
		t.Fatalf("build metadata mismatch: %#v", info)
	}
	if info.Go != runtime.Version() || info.OS != runtime.GOOS || info.Arch != runtime.GOARCH {
		t.Fatalf("runtime metadata mismatch: %#v", info)
	}
}

func TestInfoStringIncludesAllFields(t *testing.T) {
	info := Info{
		Version: "1.2.3",
		Commit:  "abcdef",
		Date:    "2026-07-13",
		Go:      "go1.99",
		OS:      "testos",
		Arch:    "testarch",
	}
	out := info.String()
	for _, want := range []string{
		"dockyard version 1.2.3",
		"commit abcdef",
		"built 2026-07-13",
		"go go1.99",
		"os/arch testos/testarch",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in version string:\n%s", want, out)
		}
	}
}
