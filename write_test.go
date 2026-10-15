package jsonedit

import "testing"

// detectColonSpace reads the first colon outside a string. A colon
// inside a key or a value must not influence the result.
func TestDetectColonSpace(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{`{}`, true},
		{`[]`, true},
		{`[1, 2, 3]`, true},
		{`{"a": 1}`, true},
		{`{"a":1}`, false},
		{`{"a": 1, "b": 2}`, true},
		{`{"a":1,"b":2}`, false},
		{`{"a:b": 1}`, true},
		{`{"a:b":1}`, false},
		{`{"a\"b":1}`, false},
		{`{"a\"b": 1}`, true},
		{`"only a string : here"`, true},
		{"{\n  \"a\": 1\n}", true},
		{"{\n  \"a\":1\n}", false},
	}
	for _, tc := range cases {
		if got := detectColonSpace([]byte(tc.in)); got != tc.want {
			t.Errorf("detectColonSpace(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
