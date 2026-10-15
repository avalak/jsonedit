package jsonedit

import "testing"

// detectJSONIndent returns the file's own indentation or "". The
// trailing newline is a text-file convention, not a signal that the
// author wanted a pretty container.
func TestDetectJSONIndent(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"compact object", `{"a": 1}`, ""},
		{"compact with trailing newline", "{\"a\": 1}\n", ""},
		{"empty object with newline", "{}\n", ""},
		{"empty array with newline", "[\n]\n", ""},
		{"two-space indent", "{\n  \"a\": 1\n}", "  "},
		{"four-space indent", "{\n    \"a\": 1\n}", "    "},
		{"tab indent", "{\n\t\"a\": 1\n}", "\t"},
		{"array with indent", "[\n  1,\n  2\n]", "  "},
		{"crlf with indent", "{\r\n  \"a\": 1\r\n}", "  "},
		{"nested uses first indented line", "{\n  \"a\": {\n    \"b\": 1\n  }\n}", "  "},
		{"no newline", `{"a": {"b": 1}}`, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := detectJSONIndent([]byte(tc.in))
			if got != tc.want {
				t.Fatalf("detectJSONIndent(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// lineIndentPrefix returns the leading whitespace of the line that
// contains pos.
func TestLineIndentPrefix(t *testing.T) {
	cases := []struct {
		name string
		in   string
		pos  int
		want string
	}{
		{"empty", "", 0, ""},
		{"no indent", "abc", 0, ""},
		{"pos mid-indent", "  abc", 2, "  "},
		{"pos past end of line", "  abc", 5, "  "},
		{"start of second line", "  abc\ndef", 6, ""},
		{"indented second line", "  abc\n    def", 8, "    "},
		{"tab indent", "\tabc", 1, "\t"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := lineIndentPrefix([]byte(tc.in), tc.pos)
			if got != tc.want {
				t.Fatalf("lineIndentPrefix(%q, %d) = %q, want %q",
					tc.in, tc.pos, got, tc.want)
			}
		})
	}
}

// indentLines prefixes every non-empty line except the first.
func TestIndentLines(t *testing.T) {
	cases := []struct {
		name, s, prefix, want string
	}{
		{"empty prefix is a no-op", "a\nb\nc", "", "a\nb\nc"},
		{"single line unchanged", "a", "  ", "a"},
		{"multi line", "a\nb\nc", "  ", "a\n  b\n  c"},
		{"empty line skipped", "a\n\nb", "  ", "a\n\n  b"},
		{"trailing newline keeps an empty tail", "a\nb\n", "  ", "a\n  b\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := indentLines(tc.s, tc.prefix)
			if got != tc.want {
				t.Fatalf("indentLines(%q, %q) = %q, want %q",
					tc.s, tc.prefix, got, tc.want)
			}
		})
	}
}

// serializeJSONValue emits primitives compactly and objects/arrays
// with the file's detected indentation.
func TestSerializeJSONValue(t *testing.T) {
	cases := []struct {
		name, raw, indent, prefix, want string
	}{
		{"int", "42", "", "", "42"},
		{"string", `"hi"`, "", "", `"hi"`},
		{"null", "null", "", "", "null"},
		{"true", "true", "", "", "true"},
		{"false", "false", "", "", "false"},
		{"empty object compact", "{}", "", "", "{}"},
		{"empty array compact", "[]", "", "", "[]"},
		{"object compact", `{"x": 1}`, "", "", `{"x":1}`},
		{"array compact", "[1, 2]", "", "", "[1,2]"},
		{"object pretty", `{"x": 1}`, "  ", "", "{\n  \"x\": 1\n}"},
		{"object pretty with prefix", `{"x": 1}`, "  ", "    ", "{\n      \"x\": 1\n    }"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := serializeJSONValue(tc.raw, tc.indent, tc.prefix)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tc.want {
				t.Fatalf("serializeJSONValue(%q, %q, %q) = %q, want %q",
					tc.raw, tc.indent, tc.prefix, got, tc.want)
			}
		})
	}
}

// assertJSONValue compares semantically: key order, whitespace, and
// numeric spelling do not matter; a value that is not valid JSON is
// an error.
func TestAssertJSONValue(t *testing.T) {
	cases := []struct {
		name    string
		got     string
		want    string
		wantErr bool
	}{
		{"equal ints", "1", "1", false},
		{"int vs float", "1", "1.0", false},
		{"float vs int", "1.0", "1", false},
		{"surrounding whitespace", "1", " 1 ", false},
		{"object key order", `{"a":1,"b":2}`, `{"b":2,"a":1}`, false},
		{"object whitespace", `{"a":1}`, `{ "a" : 1 }`, false},
		{"array whitespace", `[1,2]`, `[ 1 , 2 ]`, false},
		{"null equal", "null", "null", false},
		{"different int", "1", "2", true},
		{"null vs zero", "null", "0", true},
		{"got invalid JSON", "not json", "1", true},
		{"want invalid JSON", "1", "not json", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := assertJSONValue([]byte(tc.got), tc.want)
			if (err != nil) != tc.wantErr {
				t.Fatalf("assertJSONValue(%q, %q) err=%v, wantErr=%v",
					tc.got, tc.want, err, tc.wantErr)
			}
		})
	}
}
