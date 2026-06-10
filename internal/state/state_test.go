package state

import "testing"

func TestValidateReleaseName(t *testing.T) {
	if err := ValidateReleaseName("myapp-1"); err != nil {
		t.Fatalf("expected valid release name: %v", err)
	}
	if err := ValidateReleaseName("../bad"); err == nil {
		t.Fatal("expected invalid release name")
	}
}
