package jsonedit

import (
	"encoding/json"
	"testing"
)

// TestDeleteCRLFFirstMember pins lineStartIfOwnLine on a CRLF file:
// the function must return the position of the \r, so the caller
// removes both bytes of the line ending along with the item. The
// naive alternative returns the position of the \n and leaves an
// orphan \r before the next line.
func TestDeleteCRLFFirstMember(t *testing.T) {
	src := []byte("{\r\n  \"a\": 1,\r\n  \"b\": 2\r\n}\r\n")
	got, err := Delete(src, "$.a", "", false)
	if err != nil {
		t.Fatal(err)
	}
	want := "{\r\n  \"b\": 2\r\n}\r\n"
	if string(got) != want {
		t.Fatalf("got:\n%q\nwant:\n%q", got, want)
	}
	if !json.Valid(got) {
		t.Fatalf("result is not valid JSON: %q", got)
	}
}

// TestDeleteCRLFOnOwnLine is the same check for an array element at
// the start of its own line.
func TestDeleteCRLFOnOwnLine(t *testing.T) {
	src := []byte("[\r\n  1,\r\n  2\r\n]\r\n")
	got, err := Delete(src, "$[0]", "", false)
	if err != nil {
		t.Fatal(err)
	}
	want := "[\r\n  2\r\n]\r\n"
	if string(got) != want {
		t.Fatalf("got:\n%q\nwant:\n%q", got, want)
	}
	if !json.Valid(got) {
		t.Fatalf("result is not valid JSON: %q", got)
	}
}
