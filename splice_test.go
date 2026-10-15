package jsonedit

import "testing"

// splice replaces src[start:end] with newBytes. The result is a fresh
// slice; the source is never modified in place.
func TestSplice(t *testing.T) {
	cases := []struct {
		name        string
		src         string
		start, end  int
		replacement string
		want        string
	}{
		{"replace middle", "hello world", 6, 11, "there", "hello there"},
		{"delete range", "hello world", 5, 11, "", "hello"},
		{"insert at end", "hello", 5, 5, " world", "hello world"},
		{"insert at start", "world", 0, 0, "hello ", "hello world"},
		{"replace all", "abc", 0, 3, "xyz", "xyz"},
		{"empty source", "", 0, 0, "x", "x"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := splice([]byte(tc.src), tc.start, tc.end, []byte(tc.replacement))
			if string(got) != tc.want {
				t.Fatalf("splice(%q, %d, %d, %q) = %q, want %q",
					tc.src, tc.start, tc.end, tc.replacement, got, tc.want)
			}
		})
	}
}

// lineStartIfOwnLine returns the position of the line ending that
// precedes an item on its own line, or itemStart when other content
// shares the line. On CRLF, the position is the \r, not the \n.
func TestLineStartIfOwnLine(t *testing.T) {
	cases := []struct {
		name string
		src  string
		pos  int
		want int
	}{
		{"not own line", `[1, 2]`, 1, 1},
		{"own line LF", "[\n  1,\n  2\n]", 9, 6},
		{"own line CRLF", "[\r\n  1,\r\n  2\r\n]", 11, 7},
		{"tab indent", "[\n\t1\n]", 3, 1},
		{"start of file", "abc", 0, 0},
		{"no newline before", `{"a": 1}`, 5, 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := lineStartIfOwnLine([]byte(tc.src), tc.pos)
			if got != tc.want {
				t.Fatalf("lineStartIfOwnLine(%q, %d) = %d, want %d",
					tc.src, tc.pos, got, tc.want)
			}
		})
	}
}
