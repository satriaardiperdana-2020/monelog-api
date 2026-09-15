package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

type checkerFunc func(context.Context) error

func (fn checkerFunc) Ping(ctx context.Context) error {
	return fn(ctx)
}

func TestHealthReady(t *testing.T) {
	tests := []struct {
		name    string
		checker checkerFunc
		want    bool
	}{
		{
			name:    "database available",
			checker: func(context.Context) error { return nil },
			want:    true,
		},
		{
			name:    "database unavailable",
			checker: func(context.Context) error { return errors.New("connection failed") },
			want:    false,
		},
		{
			name: "readiness timeout",
			checker: func(ctx context.Context) error {
				<-ctx.Done()
				return ctx.Err()
			},
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			health := NewHealth(test.checker, 10*time.Millisecond)
			if got := health.Ready(context.Background()); got != test.want {
				t.Errorf("Ready() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestHealthLiveDoesNotCheckDatabase(t *testing.T) {
	called := false
	health := NewHealth(checkerFunc(func(context.Context) error {
		called = true
		return errors.New("unexpected call")
	}), time.Second)

	if !health.Live() {
		t.Fatal("Live() = false, want true")
	}
	if called {
		t.Fatal("Live() called the database readiness checker")
	}
}
