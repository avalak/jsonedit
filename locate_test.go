package jsonedit

import (
	"strings"
	"testing"
)

// Locate is the read-only entry point: it returns the byte range of
// the value at path without modifying the source.

func TestLocatePrimitive(t *testing.T) {
	src := []byte(`{"a":1}`)
	r, err := Locate(src, "$.a")
	if err != nil {
		t.Fatal(err)
	}
	if string(src[r.Start:r.End]) != "1" {
		t.Fatalf("range wrong: [%d,%d) = %q", r.Start, r.End, src[r.Start:r.End])
	}
}

// Locate rejects a source that is not valid JSON before walking.
func TestLocateRejectsInvalidSource(t *testing.T) {
	_, err := Locate([]byte("not json"), "$.a")
	if err == nil || !strings.Contains(err.Error(), "not valid JSON") {
		t.Fatalf("want invalid-source error, got %v", err)
	}
}
