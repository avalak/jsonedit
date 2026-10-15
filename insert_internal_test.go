package jsonedit

import "testing"

// childContainerType chooses the empty container that an auto-create
// will descend into: an object for a key step, an array for an index
// or append step.
func TestChildContainerType(t *testing.T) {
	cases := []struct {
		name string
		step jsonStep
		want string
	}{
		{"key step", jsonStep{Key: "a"}, "{}"},
		{"index step", jsonStep{IsIdx: true, Index: 0}, "[]"},
		{"append step", jsonStep{IsAppend: true}, "[]"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := string(childContainerType(tc.step))
			if got != tc.want {
				t.Fatalf("childContainerType(%+v) = %q, want %q",
					tc.step, got, tc.want)
			}
		})
	}
}
