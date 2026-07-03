package catalog

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

const (
	DefaultRegistry = "ghcr.io/nandub/dockyard-packages"
	EnvRegistry     = "DOCKYARD_CATALOG"
)

type Package struct {
	Name        string `json:"name"`
	Latest      string `json:"latest"`
	Description string `json:"description"`
}

var packages = map[string]Package{
	"adminer":     {Name: "adminer", Latest: "0.1.0", Description: "Lightweight database administration UI."},
	"alloy":       {Name: "alloy", Latest: "0.1.0", Description: "Grafana Alloy telemetry collector."},
	"caddy":       {Name: "caddy", Latest: "0.1.0", Description: "Caddy web server and reverse proxy."},
	"grafana":     {Name: "grafana", Latest: "0.1.0", Description: "Grafana dashboards and visualization."},
	"keycloak":    {Name: "keycloak", Latest: "0.1.0", Description: "Identity and access management service."},
	"loki":        {Name: "loki", Latest: "0.1.0", Description: "Grafana Loki log aggregation."},
	"mariadb":     {Name: "mariadb", Latest: "0.1.0", Description: "MariaDB relational database."},
	"meilisearch": {Name: "meilisearch", Latest: "0.1.0", Description: "Meilisearch search engine."},
	"minio":       {Name: "minio", Latest: "0.1.0", Description: "S3-compatible object storage."},
	"nats":        {Name: "nats", Latest: "0.1.0", Description: "NATS messaging server."},
	"nginx":       {Name: "nginx", Latest: "0.1.0", Description: "Nginx web server."},
	"pgadmin":     {Name: "pgadmin", Latest: "0.1.0", Description: "pgAdmin PostgreSQL administration UI."},
	"postgres":    {Name: "postgres", Latest: "0.1.0", Description: "PostgreSQL relational database."},
	"prometheus":  {Name: "prometheus", Latest: "0.1.0", Description: "Prometheus metrics server."},
	"rabbitmq":    {Name: "rabbitmq", Latest: "0.1.0", Description: "RabbitMQ message broker."},
	"redis":       {Name: "redis", Latest: "0.1.0", Description: "Redis-compatible in-memory data store."},
	"traefik":     {Name: "traefik", Latest: "0.1.0", Description: "Traefik reverse proxy."},
	"typesense":   {Name: "typesense", Latest: "0.1.0", Description: "Typesense search engine."},
	"uptime-kuma": {Name: "uptime-kuma", Latest: "0.1.0", Description: "Uptime Kuma monitoring dashboard."},
}

var packageNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

func Registry() string {
	registry := strings.TrimSpace(os.Getenv(EnvRegistry))
	if registry == "" {
		return DefaultRegistry
	}
	registry = strings.TrimPrefix(registry, "oci://")
	return strings.TrimRight(registry, "/")
}

func List() []Package {
	out := make([]Package, 0, len(packages))
	for _, pkg := range packages {
		out = append(out, pkg)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}

func Get(name string) (Package, bool) {
	pkg, ok := packages[name]
	return pkg, ok
}

func Resolve(input string) (string, bool, error) {
	if strings.HasPrefix(input, "catalog://") {
		ref := strings.TrimPrefix(input, "catalog://")
		name, version, err := parseRef(ref)
		if err != nil {
			return "", true, err
		}
		resolved, err := ResolveName(name, version)
		return resolved, true, err
	}
	if pkg, ok := packages[input]; ok {
		return formatOCI(pkg.Name, pkg.Latest), true, nil
	}
	return input, false, nil
}

func ResolveName(name string, version string) (string, error) {
	pkg, ok := packages[name]
	if !ok {
		return "", fmt.Errorf("package %q was not found in the configured catalog", name)
	}
	if version == "" {
		version = pkg.Latest
	}
	if version == "" {
		return "", fmt.Errorf("package %q has no default version; specify catalog://%s:VERSION", name, name)
	}
	return formatOCI(name, version), nil
}

func parseRef(ref string) (string, string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", "", fmt.Errorf("catalog source is missing a package name")
	}
	name := ref
	version := ""
	if i := strings.LastIndex(ref, ":"); i >= 0 {
		name = ref[:i]
		version = ref[i+1:]
		if version == "" {
			return "", "", fmt.Errorf("catalog source %q is missing a version after ':'", ref)
		}
	}
	if !packageNamePattern.MatchString(name) {
		return "", "", fmt.Errorf("invalid catalog package name %q", name)
	}
	return name, version, nil
}

func formatOCI(name string, version string) string {
	return fmt.Sprintf("oci://%s/%s:%s", Registry(), name, version)
}
