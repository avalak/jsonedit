package jsonedit

import "testing"

// FuzzParseJSONPath exercises the path parser. On success, every
// step names exactly one thing -- a key, an index, or the append
// marker -- because the walker switches on which is set. A step that
// is two of the three at once would dispatch two branches, and the
// walker would take whichever wins first.
func FuzzParseJSONPath(f *testing.F) {
	seeds := []string{
		`$`,
		`$.a`,
		`$.a.b.c`,
		`$[0]`,
		`$[10]`,
		`$.a[0].b`,
		`$[]`,
		`$.a[]`,
		`$..`,
		`$.`,
		``,
		`a`,
		`$.a[`,
		`$.a[abc]`,
		`$[-1]`,
		`$.a[[0]]`,
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		steps, err := parseJSONPath(s)
		if err != nil {
			return
		}
		for i, step := range steps {
			kinds := 0
			if step.Key != "" {
				kinds++
			}
			if step.IsIdx {
				kinds++
			}
			if step.IsAppend {
				kinds++
			}
			if kinds != 1 {
				t.Fatalf("parseJSONPath(%q): step %d has %d kinds: %+v",
					s, i, kinds, step)
			}
		}
	})
}
