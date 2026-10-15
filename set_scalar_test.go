package jsonedit

import "testing"

// Scalar replacement at a single path. A primitive on the right side
// is emitted compactly; the byte range around it is untouched.

func TestSetNull(t *testing.T) {
	got, err := Set([]byte(`{"a": 1}`), "$.a", "null", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"a": null}` {
		t.Fatalf("got %q", got)
	}
}

// false must not be dropped or treated as an empty value.
func TestSetFalse(t *testing.T) {
	got, err := Set([]byte(`{"a": true}`), "$.a", "false", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"a": false}` {
		t.Fatalf("got %q", got)
	}
}

// json.Unmarshal decodes whole numbers as float64; json.Marshal of a
// whole float64 emits it without a decimal point, so "1000000" in
// gives "1000000" out. Precision beyond 2^53 is lost, but the format
// is not.
func TestSetIntegerStaysInteger(t *testing.T) {
	got, err := Set([]byte(`{"a": 0}`), "$.a", "1000000", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"a": 1000000}` {
		t.Fatalf("got %q", got)
	}
}
