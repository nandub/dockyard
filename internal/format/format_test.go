package format

import "testing"

func TestSupportedFormatsIncludesCurrentAPIVersions(t *testing.T) {
	formats := SupportedFormats()
	if len(formats) != 4 {
		t.Fatalf("expected 4 supported formats, got %d: %#v", len(formats), formats)
	}

	want := map[string]struct {
		apiVersion string
		stability  Stability
	}{
		"Dockyard package manifest":   {apiVersion: ManifestAPIVersion, stability: Stable},
		"Dockyard lockfile":           {apiVersion: LockfileAPIVersion, stability: Experimental},
		"Dockyard package provenance": {apiVersion: ProvenanceAPIVersion, stability: Experimental},
		"Dockyard release state":      {apiVersion: ReleaseAPIVersion, stability: Experimental},
	}

	for _, got := range formats {
		expected, ok := want[got.Name]
		if !ok {
			t.Fatalf("unexpected supported format %q", got.Name)
		}
		if got.APIVersion != expected.apiVersion || got.Stability != expected.stability || got.Notes == "" {
			t.Fatalf("unexpected format metadata for %q: %#v", got.Name, got)
		}
		delete(want, got.Name)
	}
	if len(want) != 0 {
		t.Fatalf("missing supported formats: %#v", want)
	}
}
