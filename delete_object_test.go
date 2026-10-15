package jsonedit

import "testing"

// ── object members ─────────────────────────────

func TestDeleteObjectMiddle(t *testing.T) {
	got, err := Delete([]byte(`{"a": 1, "b": 2, "c": 3}`), "$.b", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"a": 1, "c": 3}` {
		t.Fatalf("got %q", got)
	}
}

func TestDeleteObjectFirst(t *testing.T) {
	got, err := Delete([]byte(`{"a": 1, "b": 2, "c": 3}`), "$.a", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"b": 2, "c": 3}` {
		t.Fatalf("got %q", got)
	}
}

func TestDeleteObjectLast(t *testing.T) {
	got, err := Delete([]byte(`{"a": 1, "b": 2, "c": 3}`), "$.c", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"a": 1, "b": 2}` {
		t.Fatalf("got %q", got)
	}
}

func TestDeleteObjectOnly(t *testing.T) {
	got, err := Delete([]byte(`{"a": 1}`), "$.a", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{}` {
		t.Fatalf("got %q", got)
	}
}

func TestDeleteObjectMultilineMiddle(t *testing.T) {
	src := []byte("{\n  \"a\": 1,\n  \"b\": 2,\n  \"c\": 3\n}\n")
	got, err := Delete(src, "$.b", "", false)
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"a\": 1,\n  \"c\": 3\n}\n"
	if string(got) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestDeleteObjectMultilineFirst(t *testing.T) {
	src := []byte("{\n  \"a\": 1,\n  \"b\": 2\n}\n")
	got, err := Delete(src, "$.a", "", false)
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"b\": 2\n}\n"
	if string(got) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestDeleteObjectMultilineLast(t *testing.T) {
	src := []byte("{\n  \"a\": 1,\n  \"b\": 2\n}\n")
	got, err := Delete(src, "$.b", "", false)
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"a\": 1\n}\n"
	if string(got) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestDeleteObjectMultilineOnly(t *testing.T) {
	src := []byte("{\n  \"a\": 1\n}\n")
	got, err := Delete(src, "$.a", "", false)
	if err != nil {
		t.Fatal(err)
	}
	want := "{}\n"
	if string(got) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestDeleteNested(t *testing.T) {
	got, err := Delete([]byte(`{"a": {"b": {"c": 1, "d": 2}, "e": 3}}`), "$.a.b.c", "", false)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"a": {"b": {"d": 2}, "e": 3}}`
	if string(got) != want {
		t.Fatalf("got %q", got)
	}
}

func TestDeletePreservesSiblings(t *testing.T) {
	src := []byte("{\n  \"a\": 1,\n  \"b\": 2,\n  \"c\": 3,\n  \"d\": 4\n}\n")
	got, err := Delete(src, "$.b", "", false)
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"a\": 1,\n  \"c\": 3,\n  \"d\": 4\n}\n"
	if string(got) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}
