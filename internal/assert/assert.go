package assert

import (
	"cmp"
	"reflect"
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
		t.Errorf("got: %#v, want: !nil", err)
	}
}

func NoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("got: %#v, want: nil", err)
	}
}

func DeepEqual(t *testing.T, got, want any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got: %#v, want: %#v", got, want)
	}
}

func DeepNotEqual(t *testing.T, got, want any) {
	t.Helper()
	if reflect.DeepEqual(got, want) {
		t.Errorf("got: %#v, want: !%#v", got, want)
	}
}

func Equal[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got: %#v, want: %#v", got, want)
	}
}

func NotEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got == want {
		t.Errorf("got: %#v, want: !%#v", got, want)
	}
}

func False(t *testing.T, got bool) {
	t.Helper()
	if got {
		t.Errorf("got: %#v, want: false", got)
	}
}

func True(t *testing.T, got bool) {
	t.Helper()
	if !got {
		t.Errorf("got: %#v, want: true", got)
	}
}

func GreaterOrEqual[T cmp.Ordered](t *testing.T, got, want T) {
	t.Helper()
	if !(got >= want) {
		t.Errorf("got: %#v, want: >=%#v", got, want)
	}
}

func Greater[T cmp.Ordered](t *testing.T, got, want T) {
	t.Helper()
	if !(got > want) {
		t.Errorf("got: %#v, want: >%#v", got, want)
	}
}

func LessOrEqual[T cmp.Ordered](t *testing.T, got, want T) {
	t.Helper()
	if !(got <= want) {
		t.Errorf("got: %#v, want: <=%#v", got, want)
	}
}

func Less[T cmp.Ordered](t *testing.T, got, want T) {
	t.Helper()
	if !(got < want) {
		t.Errorf("got: %#v, want: <%#v", got, want)
	}
}
