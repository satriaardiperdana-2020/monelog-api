package auth

import "testing"

func TestRefreshSecretsAreOpaqueAndHashed(t *testing.T) {
	first, err := NewRefreshSecret()
	if err != nil {
		t.Fatalf("NewRefreshSecret() error = %v", err)
	}
	second, err := NewRefreshSecret()
	if err != nil {
		t.Fatalf("NewRefreshSecret() error = %v", err)
	}
	if first == second || len(first) < 40 {
		t.Fatal("refresh secrets are not independently random")
	}
	if HashRefreshSecret(first) == first || HashRefreshSecret(first) == HashRefreshSecret(second) {
		t.Fatal("refresh hash was not distinct from its secret")
	}
	if !EqualSecret(first, first) || EqualSecret(first, second) {
		t.Fatal("constant-time comparison returned the wrong result")
	}
}
