package jsonedit

import (
	"errors"
	"testing"
)

// Every public function wraps one of the sentinels in errors.go.
// errors.Is discriminates the category; the message carries the
// detail. This file pins the contract so a refactor that changes
// which sentinel a function returns does not go unnoticed.

func TestSentinelContract(t *testing.T) {
	cases := []struct {
		name string
		call func() error
		want error
	}{
		{
			"Set/ErrInvalidSource",
			func() error { _, err := Set([]byte("x"), "$.a", "1", "", false); return err },
			ErrInvalidSource,
		},
		{
			"Set/ErrInvalidValue",
			func() error { _, err := Set([]byte(`{"a": 1}`), "$.a", "x", "", false); return err },
			ErrInvalidValue,
		},
		{
			"Set/ErrInvalidPath",
			func() error { _, err := Set([]byte(`{"a": 1}`), "no-dollar", "1", "", false); return err },
			ErrInvalidPath,
		},
		{
			"Set/ErrNotFound",
			func() error { _, err := Set([]byte(`{"a": 1}`), "$.b", "1", "", false); return err },
			ErrNotFound,
		},
		{
			"Set/ErrOutOfRange",
			func() error { _, err := Set([]byte(`[1, 2]`), "$[5]", "1", "", false); return err },
			ErrOutOfRange,
		},
		{
			"Set/ErrTypeMismatch",
			func() error { _, err := Set([]byte(`{"a": 1}`), "$.a.b", "1", "", false); return err },
			ErrTypeMismatch,
		},
		{
			"Set/ErrExpectMismatch",
			func() error { _, err := Set([]byte(`{"a": 1}`), "$.a", "2", "99", true); return err },
			ErrExpectMismatch,
		},

		{
			"Delete/ErrRootOp",
			func() error { _, err := Delete([]byte(`{"a": 1}`), "$", "", false); return err },
			ErrRootOp,
		},

		{
			"Insert/ErrCollision",
			func() error { _, err := Insert([]byte(`{"a": 1}`), "$.a", "1"); return err },
			ErrCollision,
		},
		{
			"Insert/ErrOutOfRange",
			func() error { _, err := Insert([]byte(`[1, 2]`), "$[5]", "1"); return err },
			ErrOutOfRange,
		},

		{
			"Append/ErrInvalidPath",
			func() error { _, err := Append([]byte(`[1]`), "$", "2"); return err },
			ErrInvalidPath,
		},

		{
			"Rename/ErrCollision",
			func() error { _, err := Rename([]byte(`{"a": 1, "b": 2}`), "$.a", "b", "", false); return err },
			ErrCollision,
		},
		{
			"Rename/ErrRootOp",
			func() error { _, err := Rename([]byte(`{"a": 1}`), "$", "b", "", false); return err },
			ErrRootOp,
		},

		{
			"Locate/ErrInvalidSource",
			func() error { _, err := Locate([]byte("x"), "$.a"); return err },
			ErrInvalidSource,
		},

		{
			"Validate/ErrInvalidSource",
			func() error { return Validate([]byte("x")) },
			ErrInvalidSource,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.call()
			if err == nil {
				t.Fatalf("want %v, got nil", tc.want)
			}
			if !errors.Is(err, tc.want) {
				t.Fatalf("errors.Is(%v, %v) = false", err, tc.want)
			}
		})
	}
}
