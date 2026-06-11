package format

type Stability string

const (
	Stable       Stability = "stable"
	Experimental Stability = "experimental"
)

const (
	ManifestAPIVersion   = "dockyard.dev/v1alpha1"
	LockfileAPIVersion   = "dockyard.dev/lockfile/v1alpha1"
	ProvenanceAPIVersion = "dockyard.dev/provenance/v1alpha1"
	ReleaseAPIVersion    = "dockyard.dev/release/v1alpha1"
)

type Format struct {
	Name       string    `json:"name"`
	APIVersion string    `json:"apiVersion"`
	Stability  Stability `json:"stability"`
	Notes      string    `json:"notes"`
}

func SupportedFormats() []Format {
	return []Format{
		{
			Name:       "Dockyard package manifest",
			APIVersion: ManifestAPIVersion,
			Stability:  Stable,
			Notes:      "Dockyard.yaml package metadata and Compose entrypoint. Stable candidate for v1.0.",
		},
		{
			Name:       "Dockyard lockfile",
			APIVersion: LockfileAPIVersion,
			Stability:  Experimental,
			Notes:      "dockyard.lock digest format; still experimental during v1.0 release-candidate testing.",
		},
		{
			Name:       "Dockyard package provenance",
			APIVersion: ProvenanceAPIVersion,
			Stability:  Experimental,
			Notes:      "package.provenance.json in .dockyard.tgz archives.",
		},
		{
			Name:       "Dockyard release state",
			APIVersion: ReleaseAPIVersion,
			Stability:  Experimental,
			Notes:      "$DOCKYARD_HOME release.json state. Experimental; v0.11+ can read legacy records without apiVersion.",
		},
	}
}
