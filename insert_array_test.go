package jsonedit

import (
	"strings"
	"testing"
)

// ── insert into arrays ──────────────────────────

func TestInsertArrayEmptyIndex0(t *testing.T) {
	got, err := Insert([]byte(`[]`), "$[0]", "42")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `[42]` {
		t.Fatalf("got %q", got)
	}
}

func TestInsertArrayHead(t *testing.T) {
	got, err := Insert([]byte(`[1, 2, 3]`), "$[0]", "0")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `[0, 1, 2, 3]` {
		t.Fatalf("got %q", got)
	}
}

func TestInsertArrayMiddle(t *testing.T) {
	got, err := Insert([]byte(`[1, 2, 3]`), "$[1]", "9")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `[1, 9, 2, 3]` {
		t.Fatalf("got %q", got)
	}
}

func TestInsertArrayTail(t *testing.T) {
	got, err := Insert([]byte(`[1, 2, 3]`), "$[3]", "4")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `[1, 2, 3, 4]` {
		t.Fatalf("got %q", got)
	}
}

func TestInsertArrayOutOfRange(t *testing.T) {
	_, err := Insert([]byte(`[1, 2, 3]`), "$[5]", "4")
	if err == nil || !strings.Contains(err.Error(), "out of range") {
		t.Fatalf("want out-of-range error, got %v", err)
	}
}

func TestInsertArrayAutoCreate(t *testing.T) {
	got, err := Insert([]byte(`{}`), "$.items[0]", "42")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"items": [42]}` {
		t.Fatalf("got %q", got)
	}
}

// ── append ────────────────────────────────────

func TestAppendEmpty(t *testing.T) {
	got, err := Append([]byte(`[]`), "$[]", "1")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `[1]` {
		t.Fatalf("got %q", got)
	}
}

func TestAppendNonEmpty(t *testing.T) {
	got, err := Append([]byte(`[1, 2, 3]`), "$[]", "4")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `[1, 2, 3, 4]` {
		t.Fatalf("got %q", got)
	}
}

func TestAppendNested(t *testing.T) {
	got, err := Append([]byte(`{"items": [1]}`), "$.items[]", "2")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"items": [1, 2]}` {
		t.Fatalf("got %q", got)
	}
}

func TestAppendAutoCreate(t *testing.T) {
	got, err := Append([]byte(`{}`), "$.items[]", "1")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"items": [1]}` {
		t.Fatalf("got %q", got)
	}
}

func TestAppendRequiresBrackets(t *testing.T) {
	_, err := Append([]byte(`{"items": []}`), "$.items", "1")
	if err == nil || !strings.Contains(err.Error(), "ending in []") {
		t.Fatalf("want append-requires-brackets error, got %v", err)
	}
}

// ── errors ────────────────────────────────────

// append does not add another element on a second call: it appends,
// so the source grows. Not idempotent by design.
func TestInsertIdempotenceOnAppend(t *testing.T) {
	first, err := Append([]byte(`[]`), "$[]", "1")
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != `[1]` {
		t.Fatalf("first pass: %q", first)
	}
	second, err := Append(first, "$[]", "1")
	if err != nil {
		t.Fatal(err)
	}
	if string(second) != `[1, 1]` {
		t.Fatalf("second append must add another element: %q", second)
	}
}

// !at $ names the root. There is no container to insert into; the
// public API rejects it.
func TestInsertRootRejected(t *testing.T) {
	_, err := Insert([]byte(`{}`), "$", "1")
	if err == nil || !strings.Contains(err.Error(), "cannot insert at root") {
		t.Fatalf("insert at $ must be rejected, got %v", err)
	}
	_, err = Append([]byte(`[1]`), "$", "2")
	if err == nil || !strings.Contains(err.Error(), "ending in []") {
		t.Fatalf("append at $ must be rejected, got %v", err)
	}
	// append at $[] is valid: it targets the root array itself.
	if _, err := Append([]byte(`[1]`), "$[]", "2"); err != nil {
		t.Fatalf("append at $[] must succeed, got %v", err)
	}
}
