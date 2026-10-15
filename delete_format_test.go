package jsonedit

import (
	"encoding/json"
	"testing"
)

// ── newline-before-comma ────────────────────────
//
// The byte range of a deleted value is widened to swallow the
// surrounding comma. The comma can be separated from the value by
// newlines and carriage returns, which are also consumed; the result
// must be valid JSON.

func TestDeleteNewlineBeforeComma(t *testing.T) {
	cases := []struct {
		name string
		src  string
		at   string
		want string
	}{
		{
			name: "first of two, newline before comma",
			src:  "{\"a\":0\n,\"\":0}",
			at:   "$.a",
			want: "{\"\":0}",
		},
		{
			name: "newline and spaces",
			src:  "{\"a\": 1\n  ,  \"b\": 2}",
			at:   "$.a",
			want: "{\"b\": 2}",
		},
		{
			name: "two newlines",
			src:  "{\"a\": 1\n\n,\"b\": 2}",
			at:   "$.a",
			want: "{\"b\": 2}",
		},
		{
			name: "carriage return",
			src:  "{\"a\": 1\r\n,\"b\": 2}",
			at:   "$.a",
			want: "{\"b\": 2}",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Delete([]byte(tc.src), tc.at, "", false)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
			if !json.Valid(got) {
				t.Fatalf("result is not valid JSON: %q", got)
			}
		})
	}
}
