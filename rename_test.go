package jsonedit

import (
	"strings"
	"testing"
)

func TestRenameBasic(t *testing.T) {
	got, err := Rename([]byte(`{"a": 1}`), "$.a", "b", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"b": 1}` {
		t.Fatalf("got %q", got)
	}
}

func TestRenameMiddleKeepsNeighbors(t *testing.T) {
	got, err := Rename([]byte(`{"a": 1, "b": 2, "c": 3}`), "$.b", "bb", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"a": 1, "bb": 2, "c": 3}` {
		t.Fatalf("got %q", got)
	}
}

// The value can be any JSON shape; it is reused byte for byte.
func TestRenameValueIsObject(t *testing.T) {
	got, err := Rename([]byte(`{"a": {"x": 1, "y": [2, 3]}}`), "$.a", "b", "", false)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"b": {"x": 1, "y": [2, 3]}}`
	if string(got) != want {
		t.Fatalf("got %q", got)
	}
}

func TestRenamePretty(t *testing.T) {
	src := []byte("{\n  \"a\": 1,\n  \"old\": 2,\n  \"b\": 3\n}\n")
	got, err := Rename(src, "$.old", "new", "", false)
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"a\": 1,\n  \"new\": 2,\n  \"b\": 3\n}\n"
	if string(got) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestRenameNested(t *testing.T) {
	got, err := Rename([]byte(`{"outer": {"inner": 42, "keep": 1}}`), "$.outer.inner", "renamed", "", false)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"outer": {"renamed": 42, "keep": 1}}`
	if string(got) != want {
		t.Fatalf("got %q", got)
	}
}

func TestRenameCollisionInSameObject(t *testing.T) {
	src := []byte(`{"a": 1, "b": 2}`)
	_, err := Rename(src, "$.a", "b", "", false)
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("want collision error, got %v", err)
	}
	if string(src) != `{"a": 1, "b": 2}` {
		t.Fatalf("file must not change on error: %q", src)
	}
}

// A key with the same name in a different object is fine: the
// collision check is local to the parent.
func TestRenameNoCollisionOtherObject(t *testing.T) {
	got, err := Rename([]byte(`{"a": {"x": 1}, "b": {"y": 2}}`), "$.a.x", "y", "", false)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"a": {"y": 1}, "b": {"y": 2}}`
	if string(got) != want {
		t.Fatalf("got %q", got)
	}
}

func TestRenameMissingKey(t *testing.T) {
	_, err := Rename([]byte(`{"a": 1}`), "$.missing", "x", "", false)
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("want not-found error, got %v", err)
	}
}

func TestRenameRejectsArray(t *testing.T) {
	_, err := Rename([]byte(`[1, 2]`), "$[0]", "x", "", false)
	if err == nil || !strings.Contains(err.Error(), "object keys") {
		t.Fatalf("want array-rejected error, got %v", err)
	}
}

func TestRenameRejectsArrayStepInPath(t *testing.T) {
	_, err := Rename([]byte(`{"a": []}`), "$.a[]", "x", "", false)
	if err == nil || !strings.Contains(err.Error(), "object keys") {
		t.Fatalf("want array-rejected error, got %v", err)
	}
}

func TestRenameExpectValuePass(t *testing.T) {
	exp := "42"
	got, err := Rename([]byte(`{"old": 42}`), "$.old", "new", exp, true)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"new": 42}` {
		t.Fatalf("got %q", got)
	}
}

func TestRenameExpectValueMismatch(t *testing.T) {
	exp := "99"
	src := []byte(`{"old": 42}`)
	_, err := Rename(src, "$.old", "new", exp, true)
	if err == nil || !strings.Contains(err.Error(), "expect-value") {
		t.Fatalf("want expect-value error, got %v", err)
	}
	if string(src) != `{"old": 42}` {
		t.Fatalf("file must not change: %q", src)
	}
}

// Names that need JSON escaping round-trip through json.Marshal.
func TestRenameNewKeyWithQuote(t *testing.T) {
	got, err := Rename([]byte(`{"a": 1}`), "$.a", `b"c`, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"b\"c": 1}` {
		t.Fatalf("got %q", got)
	}
}

func TestRenameFirstKey(t *testing.T) {
	got, err := Rename([]byte(`{"a": 1, "b": 2}`), "$.a", "first", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"first": 1, "b": 2}` {
		t.Fatalf("got %q", got)
	}
}

func TestRenameLastKey(t *testing.T) {
	got, err := Rename([]byte(`{"a": 1, "b": 2}`), "$.b", "last", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"a": 1, "last": 2}` {
		t.Fatalf("got %q", got)
	}
}

// A key containing an escaped double quote: "a\"b" in JSON source.
// findKeyStart walks back from the value and must land on the opening
// quote, not the inner (escaped) one, or rename splices the wrong
// bytes. The backslash-parity check exists for exactly this case.
func TestRenameEscapedQuoteInKey(t *testing.T) {
	got, err := Rename([]byte(`{"a\"b": 1}`), `$.a"b`, "c", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"c": 1}` {
		t.Fatalf("got %q", got)
	}
}

// A compact file with no space after the colon keeps the compact
// style in the renamed pair.
func TestRenameCompactNoColonSpace(t *testing.T) {
	got, err := Rename([]byte(`{"a":1}`), "$.a", "b", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"b":1}` {
		t.Fatalf("got %q", got)
	}
}
