package jsonedit

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

// FuzzLocate exercises the JSON walker directly. It walks a valid
// document following a path and returns the byte range of the value
// at that path. The range must lie inside the source and must itself
// be valid JSON: the walker computes positions from
// json.Decoder.InputOffset plus a forward skip over whitespace and
// punctuation, and an off-by-one in either yields a fragment that
// json.Valid rejects. A range that starts inside a key's quotes, or
// ends before the closing brace of a nested container, is exactly
// the class of bug this invariant catches.
func FuzzLocate(f *testing.F) {
	type seed struct{ src, path string }
	seeds := []seed{
		{`{}`, "$"},
		{`{"a": 1}`, "$"},
		{`{"a": 1}`, "$.a"},
		{`{"a":1}`, "$.a"},
		{`{"a": {"b": 1}}`, "$.a"},
		{`{"a": {"b": 1}}`, "$.a.b"},
		{`{"a": {"b": {"c": {"d": {"e": 1}}}}}`, "$.a.b.c.d.e"},
		{`[1, 2, 3]`, "$[0]"},
		{`[1, 2, 3]`, "$[2]"},
		{`[1, [2, 3]]`, "$[1]"},
		{`[1, [2, 3]]`, "$[1][0]"},
		{`[[1, [2, [3]]]]`, "$[0][1][1][0]"},
		{`{"a": [1, {"b": 2}]}`, "$.a[1].b"},
		{`{"a": [{"x": 1}, {"y": 2}]}`, "$.a[1].y"},
		{`null`, "$"},
		{`"string"`, "$"},
		{`42`, "$"},
		{`true`, "$"},
		{`false`, "$"},
		{`  {  "a" : 1 }  `, "$.a"},
		{"{\n  \"a\": 1\n}\n", "$.a"},
		{"{\r\n  \"a\": 1\r\n}\r\n", "$.a"},
		{"{\n\t\"a\": 1\n}\n", "$.a"},
		{`{"a\"b": 1}`, `$.a"b`},
		{`{"a\\b": 1}`, `$.a\b`},
		{`{"ключ": 1}`, "$.ключ"},
		{`{"emoji": "😀"}`, "$.emoji"},
		{`{"": 1}`, "$."},
		{`{"key with spaces": 1}`, "$.key with spaces"},
		{`{"a": null}`, "$.a"},
		{`{"a": true, "b": false}`, "$.b"},
		{`{"a": []}`, "$.a"},
		{`{"a": {}}`, "$.a"},
		{`{"a": "line\nwith\ttab"}`, "$.a"},
		// Neighbouring values: a fragment of the wrong one is still
		// valid JSON, so only the semantic check rejects it.
		{`{"a": 1, "b": 2}`, "$.a"},
		{`{"a": 1, "b": 2}`, "$.b"},
		{`{"a":1,"b":2,"c":3}`, "$.b"},
		{`{"x": [1, 2], "y": [3, 4]}`, "$.y"},
		{`[{"k": 1}, {"k": 2}]`, "$[1]"},
		{`{"a": [1, 2, 3], "b": [4, 5, 6]}`, "$.b[1]"},
	}
	for _, s := range seeds {
		f.Add(s.src, s.path)
	}
	f.Fuzz(func(t *testing.T, src, path string) {
		if !json.Valid([]byte(src)) {
			return
		}
		s, e, _, err := locateJSONItem([]byte(src), path)
		if err != nil {
			return
		}
		if s < 0 || s > e || e > len(src) {
			t.Fatalf("locateJSONItem(%q, %q) = [%d, %d), out of range for len %d",
				src, path, s, e, len(src))
		}
		if !json.Valid([]byte(src[s:e])) {
			t.Fatalf("locateJSONItem(%q, %q) = [%d, %d) = %q, not valid JSON",
				src, path, s, e, src[s:e])
		}

		// Semantic check. The fragment must decode to the value the
		// path names in the whole document. The range invariant
		// above permits a fragment of a neighbouring value -- also
		// valid JSON -- which this comparison rejects. referenceLookup
		// walks an already-decoded document with type switches; it
		// shares no code with walkJSON.
		//
		// The comparison is skipped when the document has duplicate
		// object keys. JSON does not define behavior for them, and
		// the two navigations legitimately disagree: walkJSON reads
		// tokens in order and stops at the first match, while a
		// map[string]any built by json.Unmarshal keeps only the last
		// value for a repeated key. Neither is wrong; the input is.
		// The structural checks still apply. The fuzz found this with
		// {"b":0,"b":1} at $.b.
		if hasDuplicateKeys([]byte(src)) {
			return
		}
		steps, err := parseJSONPath(path)
		if err != nil {
			return
		}
		var doc any
		if err := json.Unmarshal([]byte(src), &doc); err != nil {
			return
		}
		var got any
		if err := json.Unmarshal([]byte(src[s:e]), &got); err != nil {
			return
		}
		want, ok := referenceLookup(doc, steps)
		if !ok {
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("semantic mismatch:\n  src=%q\n  path=%q\n  range=[%d,%d)=%q\n  got=%#v\n  want=%#v",
				src, path, s, e, src[s:e], got, want)
		}
	})
}

// referenceLookup navigates doc by steps and returns the value at
// that path. used is false when the path does not name a value:
// an index into a non-array, a key in a non-object, an out-of-range
// index, or a missing key. The caller skips the semantic check on
// used=false -- the failure mode is "no such path", not "wrong
// path", and it is already covered by the range invariant.
func referenceLookup(doc any, steps []jsonStep) (any, bool) {
	cur := doc
	for _, s := range steps {
		if s.IsAppend {
			return nil, false
		}
		if s.IsIdx {
			arr, ok := cur.([]any)
			if !ok || s.Index < 0 || s.Index >= len(arr) {
				return nil, false
			}
			cur = arr[s.Index]
			continue
		}
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		v, ok := m[s.Key]
		if !ok {
			return nil, false
		}
		cur = v
	}
	return cur, true
}

// hasDuplicateKeys reports whether any object in src has the same
// key more than once. A decoder-level check: it reads tokens, so it
// sees every occurrence, unlike a map decode that would silently
// keep the last one. A malformed document returns false -- json.Valid
// has already run, and the caller only needs to know whether an
// otherwise-valid document is ambiguous under the two navigations.
func hasDuplicateKeys(src []byte) bool {
	dec := json.NewDecoder(bytes.NewReader(src))
	return dupValue(dec)
}

func dupValue(dec *json.Decoder) bool {
	tok, err := dec.Token()
	if err != nil {
		return false
	}
	d, ok := tok.(json.Delim)
	if !ok {
		return false
	}
	switch d {
	case '{':
		seen := map[string]bool{}
		for dec.More() {
			keyTok, err := dec.Token()
			if err != nil {
				return false
			}
			key, _ := keyTok.(string)
			if seen[key] {
				return true
			}
			seen[key] = true
			if dupValue(dec) {
				return true
			}
		}
		_, _ = dec.Token()
	case '[':
		for dec.More() {
			if dupValue(dec) {
				return true
			}
		}
		_, _ = dec.Token()
	}
	return false
}
