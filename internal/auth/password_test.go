package auth

import (
	"strings"
	"testing"
)

func TestPasswordHashAndPolicy(t *testing.T) {
	password := "correct horse battery staple"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$v=19$m=65536,t=3,p=1$") {
		t.Fatalf("unexpected PHC hash: %q", hash)
	}
	if !VerifyPassword(hash, password) || VerifyPassword(hash, password+"!") {
		t.Fatal("password verification did not distinguish the password")
	}
	for _, password := range []string{"short", strings.Repeat("a", 129)} {
		if err := ValidatePassword(password); err == nil {
			t.Fatalf("ValidatePassword(%q) = nil, want error", password)
		}
	}
	if VerifyPassword("$argon2id$v=19$m=1,t=1,p=1$bad$bad", "correct horse battery staple") {
		t.Fatal("verification accepted altered parameters")
	}
}
