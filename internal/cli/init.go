package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init DIRECTORY",
		Short: "Create a new Dockyard package",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := filepath.Clean(args[0])
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
			files := map[string]string{
				"Dockyard.yaml":      defaultManifest,
				"values.yaml":        defaultValues,
				"values.schema.json": defaultSchema,
				"compose.yaml":       defaultCompose,
				"README.md":          "# Dockyard Package\n",
				"SECURITY.md":        "# Security\n\nReport vulnerabilities privately to the maintainer.\n",
			}
			for name, content := range files {
				path := filepath.Join(dir, name)
				if _, err := os.Stat(path); err == nil {
					return fmt.Errorf("refusing to overwrite existing file %q", path)
				}
				if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
					return err
				}
			}
			fmt.Printf("created Dockyard package in %s\n", dir)
			return nil
		},
	}
	return cmd
}

const defaultManifest = `apiVersion: dockyard.dev/v1alpha1
name: example-app
description: Example Docker Compose application
version: 0.1.0
appVersion: "1.0.0"
type: application

compose:
  base: compose.yaml
  overlays: {}

security:
  requireNonRoot: false
  requireHealthchecks: true
  requireReadOnlyRootFilesystem: false
  requireNoNewPrivileges: true
  requireCapDropAll: false
  disallowPrivileged: true
  disallowHostNetwork: true
  disallowDockerSocketMount: true
  disallowHostPathMounts: false
  disallowLatestTag: true
`

const defaultValues = `image:
  repository: nginx
  tag: "1.27"

service:
  port: 8080
`

const defaultSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "required": ["image", "service"],
  "properties": {
    "image": {
      "type": "object",
      "required": ["repository", "tag"],
      "properties": {
        "repository": { "type": "string", "minLength": 1 },
        "tag": { "type": "string", "minLength": 1 }
      },
      "additionalProperties": false
    },
    "service": {
      "type": "object",
      "required": ["port"],
      "properties": {
        "port": { "type": "integer", "minimum": 1, "maximum": 65535 }
      },
      "additionalProperties": false
    }
  },
  "additionalProperties": false
}
`

const defaultCompose = `services:
  web:
    image: "${image.repository}:${image.tag}"
    ports:
      - "${service.port}:80"
    healthcheck:
      test: ["CMD", "nginx", "-t"]
      interval: 30s
      timeout: 5s
      retries: 3
    security_opt:
      - no-new-privileges:true
`
