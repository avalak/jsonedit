package jsonedit

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// assertJSONValue checks that the source fragment is semantically
// equal to want. Both are parsed as JSON and compared with
// reflect.DeepEqual; this ignores key order, whitespace, and numeric
// spelling ("1" and "1.0" both unmarshal to float64(1)).
func assertJSONValue(got []byte, want string) error {
	var a, b interface{}
	if err := json.Unmarshal(got, &a); err != nil {
		return fmt.Errorf("%w: current value: %v", ErrInvalidSource, err)
	}
	if err := json.Unmarshal([]byte(want), &b); err != nil {
		return fmt.Errorf("%w: expect: %v", ErrInvalidValue, err)
	}
	if !reflect.DeepEqual(a, b) {
		return fmt.Errorf("%w: want %s, got %s", ErrExpectMismatch, want, string(got))
	}
	return nil
}

// serializeJSONValue renders the new value. A primitive is emitted
// compactly. An object or array is pretty-printed with the file's
// detected indent and the current line's prefix, so a nested value
// aligns with its surroundings. Object key order in a new value is
// alphabetical; if that matters, do not use Set with an object.
func serializeJSONValue(raw, indent, linePrefix string) ([]byte, error) {
	var v interface{}
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return nil, err
	}
	switch v.(type) {
	case map[string]interface{}, []interface{}:
		var s []byte
		var err error
		if indent == "" {
			s, err = json.Marshal(v)
		} else {
			s, err = json.MarshalIndent(v, "", indent)
		}
		if err != nil {
			return nil, err
		}
		return []byte(indentLines(string(s), linePrefix)), nil
	default:
		return json.Marshal(v)
	}
}

// detectJSONIndent returns the indent of a pretty-printed file, or ""
// for a compact one. A file is pretty if some line after a newline
// begins with whitespace; the first such line's indentation is the
// answer. A file with newlines but no indented content -- `{"a": 1}\n`,
// `{}\n`, `[\n]\n` -- is compact. The trailing newline is a text-file
// convention, not a signal of pretty style, and an insertion should
// keep the shape the author chose.
func detectJSONIndent(src []byte) string {
	for i := 0; i < len(src); i++ {
		if src[i] != '\n' {
			continue
		}
		j := i + 1
		for j < len(src) && (src[j] == ' ' || src[j] == '\t') {
			j++
		}
		if j > i+1 {
			return string(src[i+1 : j])
		}
	}
	return ""
}

// lineIndentPrefix returns the leading whitespace of the line that
// contains pos.
func lineIndentPrefix(src []byte, pos int) string {
	i := pos
	for i > 0 && src[i-1] != '\n' {
		i--
	}
	j := i
	for j < len(src) && (src[j] == ' ' || src[j] == '\t') {
		j++
	}
	return string(src[i:j])
}

// indentLines prefixes every non-empty line except the first with
// prefix. The first line keeps its position because the caller is
// splicing the value into an existing column.
func indentLines(s, prefix string) string {
	if prefix == "" {
		return s
	}
	lines := strings.Split(s, "\n")
	for i := 1; i < len(lines); i++ {
		if lines[i] != "" {
			lines[i] = prefix + lines[i]
		}
	}
	return strings.Join(lines, "\n")
}

// detectColonSpace reports whether the file uses a space after the
// colon in object members. The rule drives both the colon and the
// comma in a new member: a file that writes `": "` between key and
// value uses `", "` between members, and one that writes `":"` uses
// `","`. The two styles agree in practice -- json.Marshal emits
// neither space, json.MarshalIndent and hand-written JSON emit both.
//
// The first colon outside a string decides. A document with no object
// member has no colon to observe; the function returns true, matching
// the historical behavior of Insert, Append, and Rename.
func detectColonSpace(src []byte) bool {
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
			return i+1 < len(src) && src[i+1] == ' '
		}
	}
	return true
}
