package middleware

import (
	"testing"
	"time"
)

func TestLoginRateLimiter(t *testing.T) {
	limiter := NewLoginRateLimiter()
	now := time.Now()
	limiter.now = func() time.Time { return now }
	for attempt := 0; attempt < 5; attempt++ {
		if !limiter.Allow("person@example.com", "192.0.2.4:54321") {
			t.Fatalf("attempt %d was rejected early", attempt+1)
		}
	}
	if limiter.Allow("person@example.com", "192.0.2.4:54321") {
		t.Fatal("sixth account/IP attempt was accepted")
	}
	now = now.Add(loginWindow)
	if !limiter.Allow("person@example.com", "192.0.2.4:54321") {
		t.Fatal("attempt after the rate window was rejected")
	}
}
