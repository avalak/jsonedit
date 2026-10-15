package jsonedit

import "testing"

// ── nested paths and sequential edits ──────────

// Five levels with an array in the middle, the shape of a real spec
// path like $.ops.registered.x.semantics.y.
func TestSetDeepNestedPath(t *testing.T) {
	got, err := Set([]byte(`{"a": {"b": {"c": [{"d": "y"}]}}}`),
		"$.a.b.c[0].d", `"x"`, "", false)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"a": {"b": {"c": [{"d": "x"}]}}}`
	if string(got) != want {
		t.Fatalf("got %q", got)
	}
}

// Two edits in sequence each affect only their own path.
func TestSetSequential(t *testing.T) {
	first, err := Set([]byte(`{"a": 0, "b": 0}`), "$.a", "1", "", false)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Set(first, "$.b", "2", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(second) != `{"a": 1, "b": 2}` {
		t.Fatalf("got %q", second)
	}
}

// Two keys, one whose name is a prefix of the other. The walker must
// match by full key, not by prefix.
func TestSetKeyNotPrefixMatch(t *testing.T) {
	got, err := Set([]byte(`{"a": 1, "ab": 2}`), "$.ab", "99", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"a": 1, "ab": 99}` {
		t.Fatalf("got %q", got)
	}
}
