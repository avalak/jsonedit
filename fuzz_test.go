package jsonedit

import (
	"encoding/json"
	"testing"
)

// FuzzSet exercises the public Set on valid JSON. The invariant is
// that a successful Set leaves a valid JSON document: any byte-range
// arithmetic error yields a fragment json.Valid rejects.
func FuzzSet(f *testing.F) {
	type seed struct{ src, path, value string }
	seeds := []seed{
		{`{"a": 1}`, "$.a", "2"},
		{`{"a":1}`, "$.a", "2"},
		{`{"a": 1}`, "$.a", `"x"`},
		{`{"a": 1}`, "$.a", "null"},
		{`{"a": 1}`, "$.a", "true"},
		{`{"a": 1}`, "$.a", "false"},
		{`{"a": 1}`, "$.a", `{"x": 1}`},
		{`{"a": 1}`, "$.a", `[1, 2, 3]`},
		{`[1, 2, 3]`, "$[1]", "9"},
		{`[1,2,3]`, "$[1]", "9"},
		{`{"a": {"b": 1}}`, "$.a.b", "2"},
		{`{"a": {"b": {"c": 1}}}`, "$.a.b.c", "2"},
		{`{"a": {"b": {"c": {"d": {"e": 1}}}}}`, "$.a.b.c.d.e", "2"},
		{`{"items": [{"x": 1}]}`, "$.items[0].x", "2"},
		{`{"items": [{"x": 1}, {"x": 2}]}`, "$.items[1].x", "3"},
		{`[{"a": 1}, {"b": 2}]`, "$[1].b", "3"},
		{`{"a": [1, 2, 3]}`, "$.a[0]", "9"},
		{`{"a": []}`, "$.a", `[1, 2]`},
		{`{"a": {}}`, "$.a", `{"x": 1}`},
		{`{"a": 1, "b": 2, "c": 3}`, "$.b", "99"},
		{"{\n  \"a\": 1\n}\n", "$.a", "2"},
		{"{\n    \"a\": 1\n}\n", "$.a", "2"},
		{"{\n\t\"a\": 1\n}\n", "$.a", "2"},
		{"{\r\n  \"a\": 1\r\n}\r\n", "$.a", "2"},
		{"{\"a\":1}", "$.a", "2"},
		{"{\"a\" : 1}", "$.a", "2"},
		{`{"ключ": "значение"}`, "$.ключ", `"новое"`},
		{`{"emoji": "😀"}`, "$.emoji", `"😎"`},
		{`{"a": "line\nwith\ttab"}`, "$.a", `"clean"`},
		{`{"a": "escaped \" quote"}`, "$.a", `"x"`},
		{`{"a": "back\\slash"}`, "$.a", `"x"`},
		{`{"key with spaces": 1}`, "$.key with spaces", "2"},
		{`{"a": null, "b": true, "c": [false]}`, "$.c[0]", "true"},
	}
	for _, s := range seeds {
		f.Add(s.src, s.path, s.value)
	}
	f.Fuzz(func(t *testing.T, src, path, value string) {
		if !json.Valid([]byte(src)) {
			return
		}
		if !json.Valid([]byte(value)) {
			return
		}
		out, err := Set([]byte(src), path, value, "", false)
		if err != nil {
			return
		}
		if !json.Valid(out) {
			t.Fatalf("Set result not valid JSON:\n  src=%q\n  path=%q\n  value=%q\n  out=%q",
				src, path, value, out)
		}
	})
}

// FuzzDelete exercises the public Delete on valid JSON. Two
// invariants: the result is valid JSON, and the document shrinks.
// Removing a value always takes at least its own byte plus one
// separator, so a successful Delete cannot produce a document of the
// same size or larger.
func FuzzDelete(f *testing.F) {
	type seed struct{ src, path string }
	seeds := []seed{
		{`{"a": 1}`, "$.a"},
		{`{"a":1}`, "$.a"},
		{`{"a": 1, "b": 2}`, "$.a"},
		{`{"a": 1, "b": 2}`, "$.b"},
		{`{"a":1,"b":2,"c":3}`, "$.b"},
		{`[1, 2, 3]`, "$[0]"},
		{`[1, 2, 3]`, "$[2]"},
		{`[1,2,3]`, "$[1]"},
		{`[{"a": 1}, {"b": 2}, {"c": 3}]`, "$[1]"},
		{`{"a": {"b": {"c": 1}}}`, "$.a.b.c"},
		{`{"a": {"b": {"c": 1, "d": 2}, "e": 3}}`, "$.a.b.d"},
		{`{"items": [{"x": 1}, {"x": 2}]}`, "$.items[0].x"},
		{`{"items": [{"x": 1, "y": 2}, {"x": 3}]}`, "$.items[0].y"},
		{"{\n  \"a\": 1,\n  \"b\": 2\n}\n", "$.a"},
		{"{\n  \"a\": 1,\n  \"b\": 2\n}\n", "$.b"},
		{"{\r\n  \"a\": 1,\r\n  \"b\": 2\r\n}\r\n", "$.a"},
		{"{\r\n  \"a\": 1,\r\n  \"b\": 2\r\n}\r\n", "$.b"},
		{"{\n\t\"a\": 1,\n\t\"b\": 2\n}\n", "$.a"},
		{"{\"a\":0\n,\"b\":0}", "$.a"},
		{"{\"a\": 1\n  ,  \"b\": 2}", "$.a"},
		{`{"ключ": 1, "другой": 2}`, "$.ключ"},
		{`{"emoji": 1, "b": 2}`, "$.emoji"},
		{`{"": 1, "b": 2}`, "$."},
		{`{"a\\b": 1, "c": 2}`, `$.a\b`},
	}
	for _, s := range seeds {
		f.Add(s.src, s.path)
	}
	f.Fuzz(func(t *testing.T, src, path string) {
		if !json.Valid([]byte(src)) {
			return
		}
		out, err := Delete([]byte(src), path, "", false)
		if err != nil {
			return
		}
		if !json.Valid(out) {
			t.Fatalf("Delete result not valid JSON:\n  src=%q\n  path=%q\n  out=%q",
				src, path, out)
		}
		if len(out) >= len(src) {
			t.Fatalf("Delete must shrink the document:\n  src=%d bytes\n  out=%d bytes\n  src=%q\n  path=%q",
				len(src), len(out), src, path)
		}
	})
}

// FuzzRename exercises the public Rename on valid JSON. Two
// invariants. First, a successful Rename leaves valid JSON: the new
// key is escaped through json.Marshal and the value's bytes are
// reused unchanged, so any byte-math error in the splice shows up as
// malformed output. Second, the file's colon style is preserved: the
// first colon outside a string in src is the same byte sequence in
// out, so its trailing space (or its absence) is unchanged. The
// second check only fires when src contains a colon outside a string;
// an array or an empty object has no style to preserve.
func FuzzRename(f *testing.F) {
	type seed struct{ src, path, newKey string }
	seeds := []seed{
		{`{"a": 1}`, "$.a", "b"},
		{`{"a":1}`, "$.a", "b"},
		{`{"a": 1, "b": 2}`, "$.a", "x"},
		{`{"a": 1, "b": 2}`, "$.b", "x"},
		{`{"a":1,"b":2}`, "$.b", "x"},
		{`{"outer": {"inner": 1}}`, "$.outer.inner", "renamed"},
		{`{"a": {"x": 1, "y": [2, 3]}}`, "$.a", "b"},
		{`{"items": [{"name": "x"}, {"name": "y"}]}`, "$.items[0].name", "label"},
		{"{\n  \"old\": 1\n}\n", "$.old", "new"},
		{"{\r\n  \"old\": 1\r\n}\r\n", "$.old", "new"},
		{"{\n\t\"old\": 1\n}\n", "$.old", "new"},
		{`{"a\"b": 1}`, `$.a"b`, "c"},
		{`{"ключ": 1}`, "$.ключ", "key"},
		{`{"": 1}`, "$.", "empty"},
		{`{"key with spaces": 1}`, "$.key with spaces", "squashed"},
	}
	for _, s := range seeds {
		f.Add(s.src, s.path, s.newKey)
	}
	f.Fuzz(func(t *testing.T, src, path, newKey string) {
		if !json.Valid([]byte(src)) {
			return
		}
		out, err := Rename([]byte(src), path, newKey, "", false)
		if err != nil {
			return
		}
		if !json.Valid(out) {
			t.Fatalf("Rename result not valid JSON:\n  src=%q\n  path=%q\n  newKey=%q\n  out=%q",
				src, path, newKey, out)
		}
		if hasColonOutsideString([]byte(src)) &&
			detectColonSpace([]byte(src)) != detectColonSpace(out) {
			t.Fatalf("Rename must keep the colon style:\n  src=%q\n  path=%q\n  newKey=%q\n  out=%q",
				src, path, newKey, out)
		}
	})
}

// FuzzInsert exercises the public Insert on valid JSON. Three
// invariants: the result is valid JSON, the document grows, and the
// file's colon style is preserved (same reasoning as FuzzRename).
func FuzzInsert(f *testing.F) {
	type seed struct{ src, path, value string }
	seeds := []seed{
		{`{}`, "$.a", "1"},
		{`{}`, "$.a", `{"x": 1}`},
		{`{"a": 1}`, "$.b", "2"},
		{`{"a":1}`, "$.b", "2"},
		{`{"a":1,"b":2}`, "$.c", "3"},
		{`{"a": 1, "b": 2}`, "$.c", "3"},
		{`[]`, "$[0]", "1"},
		{`[1, 2, 3]`, "$[0]", "0"},
		{`[1,2,3]`, "$[0]", "0"},
		{`[1, 2, 3]`, "$[3]", "4"},
		{`[1, 2, 3]`, "$[]", "4"},
		{`{"a": {"b": 1}}`, "$.a.c", "2"},
		{`{}`, "$.a.b.c", "1"},
		{`{"a": {}}`, "$.a.b", "1"},
		{`{"a": {"b": {}}}`, "$.a.b.c", "1"},
		{`{"items": []}`, "$.items[0]", "1"},
		{`{"items": [1, 2]}`, "$.items[2]", "3"},
		{"{\n  \"a\": 1\n}\n", "$.b", "2"},
		{"{\r\n  \"a\": 1\r\n}\r\n", "$.b", "2"},
		{"{\n\t\"a\": 1\n}\n", "$.b", "2"},
		{`{"a":1}`, "$.b", `{"x":1}`},
		{`{"a": 1}`, "$.b", `{"x": 1}`},
		{`{"ключ": 1}`, "$.другой", "2"},
	}
	for _, s := range seeds {
		f.Add(s.src, s.path, s.value)
	}
	f.Fuzz(func(t *testing.T, src, path, value string) {
		if !json.Valid([]byte(src)) {
			return
		}
		if !json.Valid([]byte(value)) {
			return
		}
		out, err := Insert([]byte(src), path, value)
		if err != nil {
			return
		}
		if !json.Valid(out) {
			t.Fatalf("Insert result not valid JSON:\n  src=%q\n  path=%q\n  value=%q\n  out=%q",
				src, path, value, out)
		}
		if len(out) <= len(src) {
			t.Fatalf("Insert must grow the document:\n  src=%d bytes\n  out=%d bytes\n  src=%q\n  path=%q\n  value=%q",
				len(src), len(out), src, path, value)
		}
		if hasColonOutsideString([]byte(src)) &&
			detectColonSpace([]byte(src)) != detectColonSpace(out) {
			t.Fatalf("Insert must keep the colon style:\n  src=%q\n  path=%q\n  value=%q\n  out=%q",
				src, path, value, out)
		}
	})
}

// FuzzAppend exercises the public Append on valid JSON. Append goes
// through the same insertIn as Insert; its path must end in []. Same
// invariants: valid JSON, growth, colon style preserved.
func FuzzAppend(f *testing.F) {
	type seed struct{ src, path, value string }
	seeds := []seed{
		{`[]`, "$[]", "1"},
		{`[]`, "$[]", `"x"`},
		{`[]`, "$[]", "null"},
		{`[]`, "$[]", `{"a": 1}`},
		{`[1]`, "$[]", "2"},
		{`[1, 2, 3]`, "$[]", "4"},
		{`[1,2,3]`, "$[]", "4"},
		{`[{"a": 1}]`, "$[]", "2"},
		{`[{"a":1}]`, "$[]", "2"},
		{`{"items": []}`, "$.items[]", "1"},
		{`{"items": [1]}`, "$.items[]", "2"},
		{`{"items": [1,2]}`, "$.items[]", "3"},
		{`{}`, "$.items[]", "1"},
		{`{}`, "$.a[]", "[]"},
		{`{"tags": ["a", "b"]}`, "$.tags[]", `"c"`},
		{"[]\n", "$[]", "1"},
		{"[\r\n  1\r\n]\r\n", "$[]", "2"},
		{"{\n  \"a\": []\n}\n", "$.a[]", "1"},
		{"{\r\n  \"a\": []\r\n}\r\n", "$.a[]", "1"},
		{`{"ключи": []}`, "$.ключи[]", "1"},
	}
	for _, s := range seeds {
		f.Add(s.src, s.path, s.value)
	}
	f.Fuzz(func(t *testing.T, src, path, value string) {
		if !json.Valid([]byte(src)) {
			return
		}
		if !json.Valid([]byte(value)) {
			return
		}
		out, err := Append([]byte(src), path, value)
		if err != nil {
			return
		}
		if !json.Valid(out) {
			t.Fatalf("Append result not valid JSON:\n  src=%q\n  path=%q\n  value=%q\n  out=%q",
				src, path, value, out)
		}
		if len(out) <= len(src) {
			t.Fatalf("Append must grow the document:\n  src=%d bytes\n  out=%d bytes\n  src=%q\n  path=%q\n  value=%q",
				len(src), len(out), src, path, value)
		}
		if hasColonOutsideString([]byte(src)) &&
			detectColonSpace([]byte(src)) != detectColonSpace(out) {
			t.Fatalf("Append must keep the colon style:\n  src=%q\n  path=%q\n  value=%q\n  out=%q",
				src, path, value, out)
		}
	})
}

// hasColonOutsideString reports whether src contains a ':' outside a
// string literal. The style invariant on Insert, Rename, and Append
// is checked only when the source has one: a document with no colon
// -- an array, an empty object -- has no style to preserve, and
// detectColonSpace returns its default.
func hasColonOutsideString(src []byte) bool {
	inString := false
	for i := 0; i < len(src); i++ {
		c := src[i]
		if inString {
			if c == '\\' && i+1 < len(src) {
				i++
				continue
			}
			if c == '"' {
				inString = false
			}
			continue
		}
		switch c {
		case '"':
			inString = true
		case ':':
			return true
		}
	}
	return false
}
