package lock

import "testing"

func TestExtractImages(t *testing.T) {
	images, err := ExtractImages([]byte(`services:
  web:
    image: nginx:1.27
  db:
    image: postgres@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(images) != 2 {
		t.Fatalf("expected 2 images, got %d", len(images))
	}
	if images[0].Service != "db" {
		t.Fatalf("expected sorted services")
	}
	if images[0].Digest == "" {
		t.Fatalf("expected digest extraction")
	}
}
