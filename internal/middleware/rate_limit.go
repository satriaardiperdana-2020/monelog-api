package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"sync"
	"time"
)

const loginWindow = 15 * time.Minute
const maxLoginBuckets = 10_000

type loginBucket struct {
	started time.Time
	count   int
}

// LoginRateLimiter keeps bounded login-attempt windows keyed by hashes, never raw email addresses.
type LoginRateLimiter struct {
	mu      sync.Mutex
	now     func() time.Time
	emailIP map[string]loginBucket
	ip      map[string]loginBucket
}

func NewLoginRateLimiter() *LoginRateLimiter {
	return &LoginRateLimiter{now: time.Now, emailIP: make(map[string]loginBucket), ip: make(map[string]loginBucket)}
}

// Allow records one login attempt and reports whether both account and IP limits permit it.
func (l *LoginRateLimiter) Allow(email, remoteAddr string) bool {
	now := l.now()
	ip := remoteIP(remoteAddr)
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cleanup(now)
	ipKey := hashKey(ip)
	emailIPKey := hashKey(email + "\x00" + ip)
	if _, ok := l.ip[ipKey]; !ok && len(l.ip) >= maxLoginBuckets {
		return false
	}
	if _, ok := l.emailIP[emailIPKey]; !ok && len(l.emailIP) >= maxLoginBuckets {
		return false
	}
	allowedIP := take(l.ip, ipKey, now, 20)
	allowedEmailIP := take(l.emailIP, emailIPKey, now, 5)
	return allowedIP && allowedEmailIP
}

func take(buckets map[string]loginBucket, key string, now time.Time, limit int) bool {
	bucket := buckets[key]
	if bucket.started.IsZero() || now.Sub(bucket.started) >= loginWindow {
		bucket = loginBucket{started: now}
	}
	if bucket.count >= limit {
		buckets[key] = bucket
		return false
	}
	bucket.count++
	buckets[key] = bucket
	return true
}

func (l *LoginRateLimiter) cleanup(now time.Time) {
	for key, bucket := range l.emailIP {
		if now.Sub(bucket.started) >= loginWindow {
			delete(l.emailIP, key)
		}
	}
	for key, bucket := range l.ip {
		if now.Sub(bucket.started) >= loginWindow {
			delete(l.ip, key)
		}
	}
}

func remoteIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		return host
	}
	return remoteAddr
}

func hashKey(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
