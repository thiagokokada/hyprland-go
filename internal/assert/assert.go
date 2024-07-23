package assert

import (
	"testing"
)

func Must1[T any](v T, err error) T {
	Must(err)
	return v
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Error(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("got: %v, want: !nil", err)
	}
}

func NoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("got: %v, want: nil", err)
	}
}

func Equal[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("got: %v, want: %v", got, want)
	}
}

func True(t *testing.T, got bool) {
	t.Helper()
	if got {
		t.Fatalf("got: %v, want: false", got)
	}
}
