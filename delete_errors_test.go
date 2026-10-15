package jsonedit

import (
	"strings"
	"testing"
)

// ── expect-value ───────────────────────────────

func TestDeleteExpectValuePass(t *testing.T) {
	exp := "1"
	got, err := Delete([]byte(`{"a": 1, "b": 2}`), "$.a", exp, true)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"b": 2}` {
		t.Fatalf("got %q", got)
	}
}

func TestDeleteExpectValueMismatch(t *testing.T) {
	exp := "99"
	src := []byte(`{"a": 1, "b": 2}`)
	_, err := Delete(src, "$.a", exp, true)
	if err == nil || !strings.Contains(err.Error(), "expect-value") {
		t.Fatalf("want expect-value error, got %v", err)
	}
	if string(src) != `{"a": 1, "b": 2}` {
		t.Fatalf("source must not change on failed assertion: %q", src)
	}
}

// ── error paths ────────────────────────────────

func TestDeletePathNotFound(t *testing.T) {
	_, err := Delete([]byte(`{"a": 1}`), "$.missing", "", false)
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("want not-found error, got %v", err)
	}
}

// !at $ names the root. Deleting the whole document is almost never
// what the author meant, and if it were, write-file is the right op.
func TestDeleteRootRejected(t *testing.T) {
	src := []byte(`{"a": 1}`)
	_, err := Delete(src, "$", "", false)
	if err == nil || !strings.Contains(err.Error(), "cannot delete the root") {
		t.Fatalf("want root-delete error, got %v", err)
	}
	if string(src) != `{"a": 1}` {
		t.Fatalf("source must not change: %q", src)
	}
}

// parseJSONPath trims surrounding whitespace, so "$", "$ " and " $"
// name the same path: the root. The guard must compare the parsed
// path, not the raw string, or a padded at clears the whole document.
// Found by FuzzStructOp with src={} at="$ ".
func TestDeleteRootWhitespace(t *testing.T) {
	for _, at := range []string{"$", "$ ", " $", "  $  "} {
		t.Run(at, func(t *testing.T) {
			src := []byte(`{"a": 1}`)
			_, err := Delete(src, at, "", false)
			if err == nil || !strings.Contains(err.Error(), "cannot delete the root") {
				t.Fatalf("at=%q must be rejected as the root, got %v", at, err)
			}
			if string(src) != `{"a": 1}` {
				t.Fatalf("source must not change: %q", src)
			}
		})
	}
}
