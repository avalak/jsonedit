package jsonedit

import (
	"encoding/json"
	"strings"
	"testing"
)

// ── insert into objects ──────────────────────────

func TestInsertObjectEmpty(t *testing.T) {
	got, err := Insert([]byte(`{}`), "$.a", "1")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"a": 1}` {
		t.Fatalf("got %q", got)
	}
}

func TestInsertObjectAppend(t *testing.T) {
	got, err := Insert([]byte(`{"a": 1, "b": 2}`), "$.c", "3")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"a": 1, "b": 2, "c": 3}` {
		t.Fatalf("got %q", got)
	}
}

// insert on an existing key is an explicit error, never a silent
// overwrite. Replace is Set; insert never replaces.
func TestInsertObjectExistingKeyError(t *testing.T) {
	src := []byte(`{"a": 1}`)
	_, err := Insert(src, "$.a", "9")
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("want key-exists error, got %v", err)
	}
	if string(src) != `{"a": 1}` {
		t.Fatalf("source must not change on error: %q", src)
	}
}

func TestInsertNestedAutoCreate(t *testing.T) {
	got, err := Insert([]byte(`{}`), "$.a.b.c", "1")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"a": {"b": {"c": 1}}}` {
		t.Fatalf("got %q", got)
	}
}

func TestInsertNestedPartiallyExisting(t *testing.T) {
	got, err := Insert([]byte(`{"a": {}}`), "$.a.b.c", "1")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"a": {"b": {"c": 1}}}` {
		t.Fatalf("got %q", got)
	}
}

func TestInsertNestedIntoExisting(t *testing.T) {
	got, err := Insert([]byte(`{"a": {"b": {}}}`), "$.a.b.c", "1")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"a": {"b": {"c": 1}}}` {
		t.Fatalf("got %q", got)
	}
}

// ── errors ────────────────────────────────────

func TestInsertWrongContainerType(t *testing.T) {
	_, err := Insert([]byte(`{"a": 5}`), "$.a.b", "1")
	if err == nil || !strings.Contains(err.Error(), "expected") {
		t.Fatalf("want type error, got %v", err)
	}
}

// Regression: FuzzInsert found that an object whose only member's
// value is a very long integer was spliced at the wrong offset. The
// decoder rejects such a value in Token() after consuming part of
// the source, and the old lastObjectMemberValueEnd returned 0 for
// that case, sending addMemberToObject to the top of the object.
// The function now returns an error; Insert does not silently
// produce malformed JSON.
func TestInsertLargeNumberValue(t *testing.T) {
	const bigNumber = "200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	src := []byte(`{"": ` + bigNumber + `}`)
	got, err := Insert(src, "$.0", "0")
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(got) {
		t.Fatalf("result not valid JSON: %q", got)
	}
	if !strings.Contains(string(got), `"0": 0`) {
		t.Fatalf("expected the new member in the result, got %q", got)
	}
}
