package jsonedit

import "testing"

// ── formatting preservation ──────────────────────

func TestInsertPretty(t *testing.T) {
	src := []byte("{\n  \"a\": 1\n}\n")
	got, err := Insert(src, "$.b", "2")
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"a\": 1,\n  \"b\": 2\n}\n"
	if string(got) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

// An empty file with a trailing newline is compact: the newline is a
// text-file convention, not a signal that the author wanted a pretty
// container. The insertion keeps the compact shape.
func TestInsertCompactEmptyWithNewline(t *testing.T) {
	got, err := Insert([]byte("{}\n"), "$.a", "1")
	if err != nil {
		t.Fatal(err)
	}
	want := "{\"a\": 1}\n"
	if string(got) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

// A compact file with one member and a trailing newline stays compact
// when a second member is added. The file's colon-space style decides
// the separators of the new member.
func TestInsertCompactWithTrailingNewline(t *testing.T) {
	got, err := Insert([]byte("{\"a\": 1}\n"), "$.b", "2")
	if err != nil {
		t.Fatal(err)
	}
	want := "{\"a\": 1, \"b\": 2}\n"
	if string(got) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

// A compact file with no space after the colon uses `":"` and `","`
// for the new member, matching the file's own style.
func TestInsertCompactNoColonSpace(t *testing.T) {
	got, err := Insert([]byte(`{"a":1}`), "$.b", "2")
	if err != nil {
		t.Fatal(err)
	}
	want := `{"a":1,"b":2}`
	if string(got) != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// With no colon to observe -- an empty object -- the spaced form is
// the default.
func TestInsertEmptyObjectDefaultStyle(t *testing.T) {
	got, err := Insert([]byte(`{}`), "$.a", "1")
	if err != nil {
		t.Fatal(err)
	}
	want := `{"a": 1}`
	if string(got) != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// An array has no colon; the compact style defaults to space between
// elements.
func TestInsertArrayDefaultStyle(t *testing.T) {
	got, err := Insert([]byte(`[1,2]`), "$[]", "3")
	if err != nil {
		t.Fatal(err)
	}
	want := `[1,2, 3]`
	if string(got) != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// set of an object value into a compact file must not pretty-print
// the new object either.
func TestSetObjectInCompactWithNewline(t *testing.T) {
	got, err := Set([]byte("{\"a\": 1}\n"), "$.a", `{"x": 1}`, "", false)
	if err != nil {
		t.Fatal(err)
	}
	// json.Marshal emits compact JSON: no space after the colon.
	want := "{\"a\": {\"x\":1}}\n"
	if string(got) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

// append to a compact array with a trailing newline.
func TestAppendCompactWithTrailingNewline(t *testing.T) {
	got, err := Append([]byte("[1, 2]\n"), "$[]", "3")
	if err != nil {
		t.Fatal(err)
	}
	want := "[1, 2, 3]\n"
	if string(got) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestAppendPretty(t *testing.T) {
	src := []byte("[\n  1,\n  2\n]\n")
	got, err := Append(src, "$[]", "3")
	if err != nil {
		t.Fatal(err)
	}
	want := "[\n  1,\n  2,\n  3\n]\n"
	if string(got) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

// Nested auto-create in a compact file stays compact; the newly
// created containers are written on the same line.
func TestInsertCompactNestedAutoCreate(t *testing.T) {
	got, err := Insert([]byte("{}\n"), "$.a.b", "1")
	if err != nil {
		t.Fatal(err)
	}
	want := "{\"a\": {\"b\": 1}}\n"
	if string(got) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestInsertCompactNewObjectValue(t *testing.T) {
	got, err := Insert([]byte("{}\n"), "$.a", `{"x": 1}`)
	if err != nil {
		t.Fatal(err)
	}
	want := "{\"a\": {\"x\":1}}\n"
	if string(got) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestInsertMultilineSiblings(t *testing.T) {
	src := []byte("{\n  \"a\": 1,\n  \"b\": 2\n}\n")
	got, err := Insert(src, "$.c", "3")
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"a\": 1,\n  \"b\": 2,\n  \"c\": 3\n}\n"
	if string(got) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestInsertArrayPrettyMiddle(t *testing.T) {
	src := []byte("[\n  1,\n  2\n]\n")
	got, err := Insert(src, "$[1]", "9")
	if err != nil {
		t.Fatal(err)
	}
	want := "[\n  1,\n  9,\n  2\n]\n"
	if string(got) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}
