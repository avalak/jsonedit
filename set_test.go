package jsonedit

import (
	"strings"
	"testing"
)

// Set: primitive, string, nested key, array element, formatting,
// expect-value, error paths.

func TestSetPrimitive(t *testing.T) {
	got, err := Set([]byte(`{"a": 1}`), "$.a", "2", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"a": 2}` {
		t.Fatalf("got %q", got)
	}
}

func TestSetString(t *testing.T) {
	got, err := Set([]byte(`{"greeting": "hi"}`), "$.greeting", `"hello"`, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"greeting": "hello"}` {
		t.Fatalf("got %q", got)
	}
}

func TestSetNestedKey(t *testing.T) {
	got, err := Set([]byte(`{"a": {"b": {"c": false}}}`), "$.a.b.c", "true", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"a": {"b": {"c": true}}}` {
		t.Fatalf("got %q", got)
	}
}

func TestSetArrayElement(t *testing.T) {
	got, err := Set([]byte(`[1, 2, 3]`), "$[1]", "99", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `[1, 99, 3]` {
		t.Fatalf("got %q", got)
	}
}

func TestSetInsideArray(t *testing.T) {
	got, err := Set([]byte(`{"items": [{"name": "y"}, {"name": "z"}]}`),
		"$.items[0].name", `"x"`, "", false)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"items": [{"name": "x"}, {"name": "z"}]}`
	if string(got) != want {
		t.Fatalf("got %q", got)
	}
}

// The splice must not touch whitespace or key order around the
// replaced value; only the value's byte range changes.
func TestSetPreservesFormatting(t *testing.T) {
	src := []byte("{\n  \"a\": 1,\n  \"b\": 2,\n  \"c\": 4\n}\n")
	got, err := Set(src, "$.b", "3", "", false)
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"a\": 1,\n  \"b\": 3,\n  \"c\": 4\n}\n"
	if string(got) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

// A new object value is pretty-printed with the file's indent and the
// current line's prefix, so the result reads like the surrounding
// document.
func TestSetNestedObject(t *testing.T) {
	src := []byte("{\n  \"a\": {}\n}\n")
	got, err := Set(src, "$.a", `{"x": 1}`, "", false)
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"a\": {\n    \"x\": 1\n  }\n}\n"
	if string(got) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

// The assertion is semantic: 1 and 1.0 both unmarshal to float64(1),
// so equal-in-value passes even when spelled differently.
func TestSetExpectValueSemantic(t *testing.T) {
	exp := "1.0"
	if _, err := Set([]byte(`{"a": 1}`), "$.a", "2", exp, true); err != nil {
		t.Fatalf("semantically equal values must pass: %v", err)
	}
}

func TestSetExpectValueMismatch(t *testing.T) {
	exp := "99"
	_, err := Set([]byte(`{"a": 1}`), "$.a", "2", exp, true)
	if err == nil || !strings.Contains(err.Error(), "expect-value") {
		t.Fatalf("want expect-value error, got %v", err)
	}
}

func TestSetKeyNotFound(t *testing.T) {
	_, err := Set([]byte(`{"a": 1}`), "$.missing", "1", "", false)
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("want not-found error, got %v", err)
	}
}

func TestSetIndexOutOfRange(t *testing.T) {
	_, err := Set([]byte(`[1, 2]`), "$[5]", "1", "", false)
	if err == nil || !strings.Contains(err.Error(), "out of range") {
		t.Fatalf("want out-of-range error, got %v", err)
	}
}

func TestSetInvalidFile(t *testing.T) {
	_, err := Set([]byte(`{not json}`), "$.a", "1", "", false)
	if err == nil || !strings.Contains(err.Error(), "not valid JSON") {
		t.Fatalf("want invalid-json error, got %v", err)
	}
}

func TestSetRejectBadPath(t *testing.T) {
	_, err := Set([]byte(`{}`), "a.b", "1", "", false)
	if err == nil || !strings.Contains(err.Error(), "must start with $") {
		t.Fatalf("want path error, got %v", err)
	}
}

func TestSetRejectBadValue(t *testing.T) {
	_, err := Set([]byte(`{"a": 1}`), "$.a", "{not json}", "", false)
	if err == nil || !strings.Contains(err.Error(), "not valid JSON") {
		t.Fatalf("want to-json error, got %v", err)
	}
}

// Setting a value to what it already is produces the same bytes.
func TestSetIdempotent(t *testing.T) {
	first, err := Set([]byte(`{"a": 2}`), "$.a", "2", "", false)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Set(first, "$.a", "2", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(second) != string(first) {
		t.Fatalf("not idempotent:\n  first:  %q\n  second: %q", first, second)
	}
}

// TestLocateDuplicateKeyFirst pins the walker's behavior on duplicate
// object keys: it returns the first match. The JSON spec does not
// define behavior for duplicates; json.Unmarshal into a map keeps the
// last. Two different choices, neither wrong; the input is. The
// FuzzLocate semantic comparison is skipped on such inputs; the
// range invariant still applies.
func TestLocateDuplicateKeyFirst(t *testing.T) {
	src := []byte(`{"b":0,"b":1}`)
	s, e, _, err := locateJSONItem(src, "$.b")
	if err != nil {
		t.Fatal(err)
	}
	if got := string(src[s:e]); got != "0" {
		t.Fatalf("locateJSONItem must return the first match, got %q", got)
	}
}
