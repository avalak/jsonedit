package jsonedit

import (
	"fmt"
	"strconv"
	"strings"
)

// jsonStep is one component of a JSON path: a key or an array index.
// At most one of the three fields is set; parseJSONPath guarantees the
// invariant, and FuzzParseJSONPath pins it.
type jsonStep struct {
	Key      string
	Index    int
	IsIdx    bool
	IsAppend bool
}

// parseJSONPath reads "$.a.b[0].c" into steps. Keys run between dots
// and brackets; indices are non-negative integers. No wildcards,
// filters, or recursion: a rich query belongs in a check, not in a
// path selector. A key may not contain '.' or '['.
func parseJSONPath(s string) ([]jsonStep, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("%w: empty", ErrInvalidPath)
	}
	if s[0] != '$' {
		return nil, fmt.Errorf("%w: must start with $, got %q", ErrInvalidPath, s)
	}
	s = s[1:]
	var steps []jsonStep
	for len(s) > 0 {
		switch s[0] {
		case '.':
			s = s[1:]
			i := 0
			for i < len(s) && s[i] != '.' && s[i] != '[' {
				i++
			}
			if i == 0 {
				return nil, fmt.Errorf("%w: empty key after '.'", ErrInvalidPath)
			}
			steps = append(steps, jsonStep{Key: s[:i]})
			s = s[i:]
		case '[':
			j := strings.IndexByte(s, ']')
			if j < 0 {
				return nil, fmt.Errorf("%w: unclosed '['", ErrInvalidPath)
			}
			inner := s[1:j]
			if inner == "" {
				steps = append(steps, jsonStep{IsAppend: true})
			} else {
				n, err := strconv.Atoi(inner)
				if err != nil {
					return nil, fmt.Errorf("%w: bad index %q: %v", ErrInvalidPath, inner, err)
				}
				if n < 0 {
					return nil, fmt.Errorf("%w: negative index %d", ErrInvalidPath, n)
				}
				steps = append(steps, jsonStep{Index: n, IsIdx: true})
			}
			s = s[j+1:]
		default:
			return nil, fmt.Errorf("%w: unexpected %q", ErrInvalidPath, string(s[0]))
		}
	}
	return steps, nil
}
