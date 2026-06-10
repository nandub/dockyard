package render

import "testing"

func TestFlattenValues(t *testing.T) {
	vals := map[string]any{"image": map[string]any{"repository": "nginx", "tag": "1.27"}}
	flat := FlattenValues("", vals)
	if flat["image.repository"] != "nginx" {
		t.Fatalf("expected image.repository to be nginx")
	}
	if flat["image.tag"] != "1.27" {
		t.Fatalf("expected image.tag to be 1.27")
	}
}

func TestIsSensitiveKey(t *testing.T) {
	if !IsSensitiveKey("database.password") {
		t.Fatal("expected password key to be sensitive")
	}
	if IsSensitiveKey("service.port") {
		t.Fatal("did not expect service.port to be sensitive")
	}
}
