package jsonedit

import "testing"

// Unicode keys and values are byte sequences like any other. The
// walker matches keys by full string; the splice preserves bytes;
// json.Marshal emits UTF-8 directly for a new key.

func TestSetUnicodeKey(t *testing.T) {
	got, err := Set([]byte(`{"ключ": 1}`), "$.ключ", "2", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"ключ": 2}` {
		t.Fatalf("got %q", got)
	}
}

func TestDeleteUnicodeKey(t *testing.T) {
	got, err := Delete([]byte(`{"ключ": 1, "другой": 2}`), "$.ключ", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"другой": 2}` {
		t.Fatalf("got %q", got)
	}
}

func TestInsertUnicodeKey(t *testing.T) {
	got, err := Insert([]byte(`{"a":1}`), "$.ключ", "2")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"a":1,"ключ":2}` {
		t.Fatalf("got %q", got)
	}
}

func TestRenameToUnicodeKey(t *testing.T) {
	got, err := Rename([]byte(`{"a": 1}`), "$.a", "ключ", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"ключ": 1}` {
		t.Fatalf("got %q", got)
	}
}

func TestLocateUnicodeKey(t *testing.T) {
	src := []byte(`{"ключ": 1}`)
	r, err := Locate(src, "$.ключ")
	if err != nil {
		t.Fatal(err)
	}
	if string(src[r.Start:r.End]) != "1" {
		t.Fatalf("range wrong: [%d,%d) = %q", r.Start, r.End, src[r.Start:r.End])
	}
}

// Emoji: four-byte UTF-8 sequences at the value position.
func TestSetEmojiValue(t *testing.T) {
	got, err := Set([]byte(`{"x": 1}`), "$.x", `"😀"`, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"x": "😀"}` {
		t.Fatalf("got %q", got)
	}
}

// A key with a space in it: the walker matches on the whole string,
// not on tokens.
func TestSetKeyWithSpace(t *testing.T) {
	got, err := Set([]byte(`{"key with spaces": 1}`), "$.key with spaces", "2", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"key with spaces": 2}` {
		t.Fatalf("got %q", got)
	}
}
