package jsonedit

import "testing"

// ── trailing newline ───────────────────────────

// A file with a trailing newline keeps it.
func TestSetTrailingNewline(t *testing.T) {
	got, err := Set([]byte("{\n  \"a\": 1\n}\n"), "$.a", "2", "", false)
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"a\": 2\n}\n"
	if string(got) != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// A file without a trailing newline stays without: the splice touches
// only the value's byte range, not the end of the file.
func TestSetNoTrailingNewline(t *testing.T) {
	got, err := Set([]byte(`{"a": 1}`), "$.a", "2", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"a": 2}` {
		t.Fatalf("got %q", got)
	}
	if got[len(got)-1] == '\n' {
		t.Fatal("a trailing newline must not be added")
	}
}

// ── indentation of a new value ─────────────────

// A file indented with four spaces produces a four-space nested
// object.
func TestSetNewObjectIndent4(t *testing.T) {
	got, err := Set([]byte("{\n    \"a\": null\n}\n"), "$.a", `{"x": 1}`, "", false)
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n    \"a\": {\n        \"x\": 1\n    }\n}\n"
	if string(got) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

// A new array inside an object.
func TestSetNewArrayInObject(t *testing.T) {
	got, err := Set([]byte("{\n  \"tags\": null\n}\n"), "$.tags", `["a", "b"]`, "", false)
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"tags\": [\n    \"a\",\n    \"b\"\n  ]\n}\n"
	if string(got) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

// A file indented with tabs produces a tab-indented nested object.
// json.MarshalIndent writes a tab after each newline; indentLines adds
// the current line's prefix to every subsequent line.
func TestSetNewObjectIndentTab(t *testing.T) {
	got, err := Set([]byte("{\n\t\"a\": null\n}\n"), "$.a", `{"x": 1}`, "", false)
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n\t\"a\": {\n\t\t\"x\": 1\n\t}\n}\n"
	if string(got) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}
