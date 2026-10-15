package jsonedit

import "testing"

// ── array elements ─────────────────────────────

func TestDeleteArrayMiddle(t *testing.T) {
	got, err := Delete([]byte(`[1, 2, 3]`), "$[1]", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `[1, 3]` {
		t.Fatalf("got %q", got)
	}
}

func TestDeleteArrayFirst(t *testing.T) {
	got, err := Delete([]byte(`[1, 2, 3]`), "$[0]", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `[2, 3]` {
		t.Fatalf("got %q", got)
	}
}

func TestDeleteArrayLast(t *testing.T) {
	got, err := Delete([]byte(`[1, 2, 3]`), "$[2]", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `[1, 2]` {
		t.Fatalf("got %q", got)
	}
}

func TestDeleteArrayOnly(t *testing.T) {
	got, err := Delete([]byte(`[42]`), "$[0]", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `[]` {
		t.Fatalf("got %q", got)
	}
}

func TestDeleteArrayElementDeep(t *testing.T) {
	got, err := Delete([]byte(`{"items": [{"name": "a", "x": 1}, {"name": "b", "x": 2}]}`),
		"$.items[1].name", "", false)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"items": [{"name": "a", "x": 1}, {"x": 2}]}`
	if string(got) != want {
		t.Fatalf("got %q", got)
	}
}
