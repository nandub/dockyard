package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

const EnvDockyardHome = "DOCKYARD_HOME"

var releaseNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,127}$`)

type Source struct {
	Type string `json:"type"`
	Path string `json:"path,omitempty"`
}

type Release struct {
	Name           string    `json:"name"`
	PackageName    string    `json:"packageName"`
	PackageVersion string    `json:"packageVersion"`
	AppVersion     string    `json:"appVersion"`
	Revision       int       `json:"revision"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	ComposeProject string    `json:"composeProject"`
	Source         Source    `json:"source"`
}

func ValidateReleaseName(name string) error {
	if !releaseNamePattern.MatchString(name) {
		return errors.New("release name must match ^[a-zA-Z0-9][a-zA-Z0-9._-]{0,127}$")
	}
	return nil
}

func Home(explicitHome string) (string, error) {
	if explicitHome != "" {
		return filepath.Abs(filepath.Clean(explicitHome))
	}
	if envHome := os.Getenv(EnvDockyardHome); envHome != "" {
		return filepath.Abs(filepath.Clean(envHome))
	}
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home: %w", err)
	}
	return filepath.Join(userHome, ".dockyard"), nil
}

func ReleaseDir(home string, releaseName string) string {
	return filepath.Join(home, "releases", releaseName)
}

func RevisionDir(home string, releaseName string, revision int) string {
	return filepath.Join(ReleaseDir(home, releaseName), "revisions", strconv.Itoa(revision))
}

func CurrentFile(home string, releaseName string) string {
	return filepath.Join(ReleaseDir(home, releaseName), "current")
}

func WriteRelease(dir string, release Release) error {
	data, err := json.MarshalIndent(release, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal release metadata: %w", err)
	}
	return os.WriteFile(filepath.Join(dir, "release.json"), append(data, '\n'), 0o600)
}

func ReadRelease(dir string) (*Release, error) {
	data, err := os.ReadFile(filepath.Join(dir, "release.json"))
	if err != nil {
		return nil, fmt.Errorf("read release metadata: %w", err)
	}
	var release Release
	if err := json.Unmarshal(data, &release); err != nil {
		return nil, fmt.Errorf("parse release metadata: %w", err)
	}
	return &release, nil
}

func ReadCurrentRevision(home string, releaseName string) (int, error) {
	data, err := os.ReadFile(CurrentFile(home, releaseName))
	if err != nil {
		return 0, fmt.Errorf("read current revision: %w", err)
	}
	revision, err := strconv.Atoi(stringTrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("parse current revision: %w", err)
	}
	return revision, nil
}

func SetCurrentRevision(home string, releaseName string, revision int) error {
	if err := os.MkdirAll(ReleaseDir(home, releaseName), 0o700); err != nil {
		return err
	}
	return os.WriteFile(CurrentFile(home, releaseName), []byte(strconv.Itoa(revision)+"\n"), 0o600)
}

func NextRevision(home string, releaseName string) (int, error) {
	revisionsDir := filepath.Join(ReleaseDir(home, releaseName), "revisions")
	entries, err := os.ReadDir(revisionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, err
	}
	maxRevision := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		revision, err := strconv.Atoi(entry.Name())
		if err == nil && revision > maxRevision {
			maxRevision = revision
		}
	}
	return maxRevision + 1, nil
}

func stringTrimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\n' || s[0] == '\t' || s[0] == '\r') {
		s = s[1:]
	}
	for len(s) > 0 {
		i := len(s) - 1
		if s[i] != ' ' && s[i] != '\n' && s[i] != '\t' && s[i] != '\r' {
			break
		}
		s = s[:i]
	}
	return s
}
