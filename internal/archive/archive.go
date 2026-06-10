package archive

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/nandub/dockyard/internal/dockpkg"
	"github.com/nandub/dockyard/internal/format"
	"github.com/nandub/dockyard/internal/lock"
)

type FileDigest struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type Provenance struct {
	APIVersion     string    `json:"apiVersion"`
	PackageName    string    `json:"packageName"`
	PackageVersion string    `json:"packageVersion"`
	AppVersion     string    `json:"appVersion"`
	BuiltAt        time.Time `json:"builtAt"`
	Builder        string    `json:"builder"`
	Locked         bool      `json:"locked"`
	LockfileSHA256 string    `json:"lockfileSha256,omitempty"`
}

var forbiddenNames = map[string]struct{}{
	".env":         {},
	"id_rsa":       {},
	"id_ed25519":   {},
	"docker.sock":  {},
	".dockyard":    {},
	".git":         {},
	"node_modules": {},
	"vendor":       {},
}

var deterministicTime = time.Unix(0, 0).UTC()

func PackageDir(packageDir string, outputFile string, requireLock bool) (string, error) {
	manifest, err := dockpkg.LoadManifest(packageDir)
	if err != nil {
		return "", err
	}
	if outputFile == "" {
		outputFile = fmt.Sprintf("%s-%s.dockyard.tgz", manifest.Name, manifest.Version)
	}

	files, err := collectFiles(packageDir)
	if err != nil {
		return "", err
	}
	files = filterGeneratedFiles(files)
	locked := false
	lockDigest := ""
	if hasFile(files, lock.FileName) {
		locked = true
		path, err := dockpkg.SafeJoin(packageDir, lock.FileName)
		if err != nil {
			return "", err
		}
		lockDigest, err = sha256File(path)
		if err != nil {
			return "", err
		}
	} else if requireLock {
		return "", fmt.Errorf("--locked requires %s; run dockyard lock first", lock.FileName)
	}

	digests, err := calculateDigests(packageDir, files)
	if err != nil {
		return "", err
	}

	provenance := Provenance{
		APIVersion:     format.ProvenanceAPIVersion,
		PackageName:    manifest.Name,
		PackageVersion: manifest.Version,
		AppVersion:     manifest.AppVersion,
		BuiltAt:        time.Now().UTC(),
		Builder:        "dockyard",
		Locked:         locked,
		LockfileSHA256: lockDigest,
	}
	provenanceBytes, err := json.MarshalIndent(provenance, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal provenance: %w", err)
	}
	provenanceBytes = append(provenanceBytes, '\n')
	digests["package.provenance.json"] = sha256Bytes(provenanceBytes)
	digestText := renderSHA256SUMS(digests)

	out, err := os.OpenFile(filepath.Clean(outputFile), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return "", fmt.Errorf("create package archive: %w", err)
	}
	defer out.Close()

	gz := gzip.NewWriter(out)
	gz.Name = ""
	gz.Comment = ""
	gz.ModTime = deterministicTime
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()

	for _, rel := range files {
		if err := addFile(tw, packageDir, rel); err != nil {
			return "", err
		}
	}
	if err := addGeneratedFile(tw, "package.provenance.json", provenanceBytes); err != nil {
		return "", err
	}
	if err := addGeneratedFile(tw, "SHA256SUMS", []byte(digestText)); err != nil {
		return "", err
	}
	return outputFile, nil
}

func VerifyArchive(archivePath string, extractLint func(tempDir string) error) error {
	tempDir, err := os.MkdirTemp("", "dockyard-verify-*")
	if err != nil {
		return fmt.Errorf("create verification temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	if err := ExtractArchive(archivePath, tempDir); err != nil {
		return err
	}

	if err := verifyForbiddenFiles(tempDir); err != nil {
		return err
	}

	sumsPath := filepath.Join(tempDir, "SHA256SUMS")
	expected, err := parseSHA256SUMS(sumsPath)
	if err != nil {
		return err
	}
	actualFiles, err := collectFiles(tempDir)
	if err != nil {
		return err
	}
	actual, err := calculateDigests(tempDir, actualFiles)
	if err != nil {
		return err
	}
	delete(actual, "SHA256SUMS")
	for path, digest := range expected {
		if actual[path] != digest {
			return fmt.Errorf("digest mismatch for %s", path)
		}
	}
	for path := range actual {
		if _, ok := expected[path]; !ok {
			return fmt.Errorf("archive contains file not listed in SHA256SUMS: %s", path)
		}
	}

	if err := verifyProvenance(tempDir); err != nil {
		return err
	}
	if _, err := dockpkg.LoadManifest(tempDir); err != nil {
		return err
	}
	if extractLint != nil {
		if err := extractLint(tempDir); err != nil {
			return err
		}
	}
	return nil
}

func ExtractArchive(archivePath string, destDir string) error {
	file, err := os.Open(filepath.Clean(archivePath))
	if err != nil {
		return fmt.Errorf("open package archive: %w", err)
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("read gzip archive: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar archive: %w", err)
		}
		if header == nil {
			continue
		}
		if header.Typeflag != tar.TypeReg {
			return fmt.Errorf("archive contains unsupported entry type for %s", header.Name)
		}
		if filepath.IsAbs(header.Name) || strings.Contains(filepath.ToSlash(header.Name), "../") || strings.Contains(header.Name, "..\\") {
			return fmt.Errorf("archive contains unsafe path %q", header.Name)
		}
		target, err := dockpkg.SafeJoin(destDir, filepath.ToSlash(header.Name))
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			return fmt.Errorf("create extracted file: %w", err)
		}
		if _, err := io.CopyN(out, tr, header.Size); err != nil {
			out.Close()
			return fmt.Errorf("extract file: %w", err)
		}
		if err := out.Close(); err != nil {
			return err
		}
	}
	return nil
}

func collectFiles(root string) ([]string, error) {
	var files []string
	rootAbs, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return nil, err
	}
	err = filepath.WalkDir(rootAbs, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		name := d.Name()
		if _, forbidden := forbiddenNames[name]; forbidden {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return fmt.Errorf("forbidden file in package: %s", name)
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlinks are not allowed in packages: %s", path)
		}
		if strings.HasSuffix(name, ".pem") || strings.HasSuffix(name, ".key") {
			return fmt.Errorf("forbidden secret-like file in package: %s", name)
		}
		rel, err := filepath.Rel(rootAbs, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "SHA256SUMS" {
			return nil
		}
		files = append(files, rel)
		return nil
	})
	sort.Strings(files)
	return files, err
}

func calculateDigests(root string, files []string) (map[string]string, error) {
	out := make(map[string]string, len(files))
	for _, rel := range files {
		path, err := dockpkg.SafeJoin(root, rel)
		if err != nil {
			return nil, err
		}
		digest, err := sha256File(path)
		if err != nil {
			return nil, err
		}
		out[rel] = digest
	}
	return out, nil
}

func renderSHA256SUMS(digests map[string]string) string {
	keys := make([]string, 0, len(digests))
	for key := range digests {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, key := range keys {
		b.WriteString(digests[key])
		b.WriteString("  ")
		b.WriteString(key)
		b.WriteByte('\n')
	}
	return b.String()
}

func parseSHA256SUMS(path string) (map[string]string, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("read SHA256SUMS: %w", err)
	}
	expected := map[string]string{}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid SHA256SUMS line")
		}
		if len(parts[0]) != 64 {
			return nil, fmt.Errorf("invalid SHA256 digest length")
		}
		expected[parts[1]] = parts[0]
	}
	return expected, nil
}

func addFile(tw *tar.Writer, root string, rel string) error {
	path, err := dockpkg.SafeJoin(root, rel)
	if err != nil {
		return err
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	header := &tar.Header{
		Name:    filepath.ToSlash(rel),
		Size:    info.Size(),
		Mode:    0o600,
		ModTime: deterministicTime,
		Uid:     0,
		Gid:     0,
		Uname:   "",
		Gname:   "",
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(tw, file)
	return err
}

func addGeneratedFile(tw *tar.Writer, name string, data []byte) error {
	header := &tar.Header{
		Name:    name,
		Size:    int64(len(data)),
		Mode:    0o600,
		ModTime: deterministicTime,
		Uid:     0,
		Gid:     0,
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	_, err := tw.Write(data)
	return err
}

func verifyForbiddenFiles(root string) error {
	_, err := collectFiles(root)
	return err
}

func verifyProvenance(root string) error {
	path := filepath.Join(root, "package.provenance.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read package provenance: %w", err)
	}
	var p Provenance
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("parse package provenance: %w", err)
	}
	if p.APIVersion != format.ProvenanceAPIVersion {
		return fmt.Errorf("unsupported provenance apiVersion %q", p.APIVersion)
	}
	if p.PackageName == "" || p.PackageVersion == "" {
		return fmt.Errorf("package provenance is missing package identity")
	}
	if p.Locked {
		digest, err := sha256File(filepath.Join(root, lock.FileName))
		if err != nil {
			return fmt.Errorf("locked package missing %s", lock.FileName)
		}
		if digest != p.LockfileSHA256 {
			return fmt.Errorf("lockfile digest mismatch in provenance")
		}
	}
	return nil
}

func hasFile(files []string, name string) bool {
	for _, file := range files {
		if file == name {
			return true
		}
	}
	return false
}

func sha256File(path string) (string, error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("open file for digest: %w", err)
	}
	defer file.Close()
	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", fmt.Errorf("hash file: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func sha256Bytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func filterGeneratedFiles(files []string) []string {
	out := make([]string, 0, len(files))
	for _, file := range files {
		if file == "SHA256SUMS" || file == "package.provenance.json" {
			continue
		}
		out = append(out, file)
	}
	return out
}
