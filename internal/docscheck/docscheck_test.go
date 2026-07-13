package docscheck

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

var requiredPlaybookSections = []string{
	"## Purpose",
	"## When to Use It",
	"## Required Reading",
	"## Preconditions",
	"## Procedure",
	"## Validation",
	"## Completion Checklist",
	"## Escalation Conditions",
	"## Required Completion Report",
}

func TestAIDocumentationReferences(t *testing.T) {
	root := repoRoot(t)

	agents := readFile(t, filepath.Join(root, "AGENTS.md"))
	for _, ref := range regexp.MustCompile("`([^`]*\\.ai/playbooks/[^`]+\\.md)`").FindAllStringSubmatch(agents, -1) {
		assertPathExists(t, root, ref[1])
	}

	for _, ref := range localPathRefs(readFile(t, filepath.Join(root, "docs", "index.md"))) {
		assertPathExists(t, root, ref)
	}
}

func TestAIPlaybooksHaveRequiredSections(t *testing.T) {
	root := repoRoot(t)
	playbooksDir := filepath.Join(root, ".ai", "playbooks")

	entries, err := os.ReadDir(playbooksDir)
	if err != nil {
		t.Fatalf("read playbooks dir: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "README.md" || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			content := readFile(t, filepath.Join(playbooksDir, entry.Name()))
			for _, section := range requiredPlaybookSections {
				if !strings.Contains(content, section) {
					t.Fatalf("missing required section %q", section)
				}
			}
		})
	}
}

func TestNoStaleAIOnboardingReferences(t *testing.T) {
	root := repoRoot(t)
	forbidden := []string{
		".agents/workflows",
		".agents/skills",
		"docs/onboarding",
		"cli-command.md",
		"Helm",
		"helm",
		"Kubernetes",
		"kubernetes",
		"chart",
		"charts",
	}

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "bin", "dist":
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}

		content := readFile(t, path)
		for _, term := range forbidden {
			if strings.Contains(content, term) {
				rel, relErr := filepath.Rel(root, path)
				if relErr != nil {
					rel = path
				}
				t.Errorf("%s contains stale or irrelevant reference %q", filepath.ToSlash(rel), term)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk repository docs: %v", err)
	}
}

func localPathRefs(markdown string) []string {
	matches := regexp.MustCompile("`([^`]+)`").FindAllStringSubmatch(markdown, -1)
	var refs []string
	for _, match := range matches {
		ref := match[1]
		if isLocalDocRef(ref) {
			refs = append(refs, ref)
		}
	}
	return refs
}

func isLocalDocRef(ref string) bool {
	prefixes := []string{
		".ai/",
		".github/",
		"docs/",
		"AGENTS.md",
		"CHANGELOG.md",
		"README.md",
		"SECURITY.md",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(ref, prefix) {
			return true
		}
	}
	return false
}

func assertPathExists(t *testing.T, root, ref string) {
	t.Helper()
	clean := strings.TrimRight(ref, "/")
	path := filepath.Join(root, filepath.FromSlash(clean))
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("referenced path %q does not exist: %v", ref, err)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
