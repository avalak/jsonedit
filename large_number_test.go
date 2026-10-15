package jsonedit

import (
	"encoding/json"
	"strings"
	"testing"
)

// bigNumber is a 300-digit integer. It is valid JSON per RFC 8259:
// json.Valid accepts it. It does not fit in a float64, and
// encoding/json's Token() rejects it with strconv.ErrRange unless
// the decoder is in UseNumber mode.
//
// jsonval uses UseNumber throughout, so the walker reads the value
// as a json.Number and continues past it. Before this was fixed,
// lastObjectMemberValueEnd returned a partial offset when the
// decoder gave up mid-value, and Insert spliced the new member into
// the middle of the existing one. FuzzInsert found the case with an
// empty key and a large number.
const bigNumber = "200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"

// A document containing a number too large for float64 is edited
// correctly. The number is preserved byte-for-byte; mutations around
// it produce valid JSON.
func TestLargeNumberMutations(t *testing.T) {
	// The exact case FuzzInsert found: an empty key plus a large
	// number. Before the fix, Insert added the new member between
	// the key's closing quote and the original member's colon,
	// producing {"","0":0:<bigNumber>}.
	t.Run("InsertPastLargeNumber", func(t *testing.T) {
		src := []byte(`{"":` + bigNumber + `}`)
		got, err := Insert(src, "$.0", "0")
		if err != nil {
			t.Fatal(err)
		}
		if !json.Valid(got) {
			t.Fatalf("Insert result not valid JSON:\n  src=%q\n  out=%q", src, got)
		}
		if !strings.Contains(string(got), `"0"`) {
			t.Fatalf("new key missing: %q", got)
		}
		if !strings.Contains(string(got), bigNumber) {
			t.Fatalf("large number must survive byte-for-byte: %q", got)
		}
	})

	t.Run("RenameLargeNumberMember", func(t *testing.T) {
		src := []byte(`{"a": ` + bigNumber + `}`)
		got, err := Rename(src, "$.a", "b", "", false)
		if err != nil {
			t.Fatal(err)
		}
		if !json.Valid(got) {
			t.Fatalf("Rename result not valid JSON: %q", got)
		}
		if !strings.Contains(string(got), bigNumber) {
			t.Fatalf("large number must survive: %q", got)
		}
	})

	t.Run("DeleteSmallMemberLeavesLargeNumber", func(t *testing.T) {
		src := []byte(`{"a": 1, "b": ` + bigNumber + `}`)
		got, err := Delete(src, "$.a", "", false)
		if err != nil {
			t.Fatal(err)
		}
		if !json.Valid(got) {
			t.Fatalf("Delete result not valid JSON: %q", got)
		}
		if !strings.Contains(string(got), bigNumber) {
			t.Fatalf("large number must survive: %q", got)
		}
	})

	t.Run("SetAnotherMemberLeavesLargeNumber", func(t *testing.T) {
		src := []byte(`{"a": 1, "b": ` + bigNumber + `}`)
		got, err := Set(src, "$.a", "2", "", false)
		if err != nil {
			t.Fatal(err)
		}
		if !json.Valid(got) {
			t.Fatalf("Set result not valid JSON: %q", got)
		}
		if !strings.Contains(string(got), bigNumber) {
			t.Fatalf("large number must survive: %q", got)
		}
	})
}
