package jsonedit

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// insertIn walks the JSON value in src, following steps. If the path
// exists along the way, it descends; if a key or an array position is
// missing, containers are created. The last step determines where the
// new value goes. prefix is the string that lines inside the current
// container should carry; colonSpace selects `": "` or `":"` for a
// new member and `", "` or `","` between members.
func insertIn(src []byte, steps []jsonStep, rawValue string, indent, prefix string, colonSpace bool) ([]byte, error) {
	if len(steps) == 0 {
		return nil, fmt.Errorf("empty steps")
	}
	step := steps[0]
	last := len(steps) == 1

	trimmed := bytes.TrimLeft(src, " \t\r\n")
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("empty source")
	}

	switch trimmed[0] {
	case '{':
		if step.IsIdx || step.IsAppend {
			return nil, fmt.Errorf("%w: expected array, got object", ErrTypeMismatch)
		}
		if last {
			if _, _, found := findMemberValueRange(src, step.Key); found {
				return nil, fmt.Errorf("%w: %q", ErrCollision, step.Key)
			}
			newValue, err := serializeJSONValue(rawValue, indent, prefix+indent)
			if err != nil {
				return nil, err
			}
			return addMemberToObject(src, step.Key, newValue, indent, prefix, colonSpace)
		}
		vs, ve, found := findMemberValueRange(src, step.Key)
		if found {
			inner, err := insertIn(src[vs:ve], steps[1:], rawValue, indent, prefix+indent, colonSpace)
			if err != nil {
				return nil, err
			}
			return splice(src, vs, ve, inner), nil
		}
		container := childContainerType(steps[1])
		inner, err := insertIn(container, steps[1:], rawValue, indent, prefix+indent, colonSpace)
		if err != nil {
			return nil, err
		}
		return addMemberToObject(src, step.Key, inner, indent, prefix, colonSpace)
	case '[':
		if !step.IsIdx && !step.IsAppend {
			return nil, fmt.Errorf("%w: expected object, got array", ErrTypeMismatch)
		}
		if last {
			newValue, err := serializeJSONValue(rawValue, indent, prefix+indent)
			if err != nil {
				return nil, err
			}
			return insertElementIntoArray(src, step, newValue, indent, prefix, colonSpace)
		}
		if step.IsAppend {
			return nil, fmt.Errorf("%w: [] must be the last path step", ErrInvalidPath)
		}
		es, ee, found := findElementRange(src, step.Index)
		if found {
			inner, err := insertIn(src[es:ee], steps[1:], rawValue, indent, prefix+indent, colonSpace)
			if err != nil {
				return nil, err
			}
			return splice(src, es, ee, inner), nil
		}
		n, err := countElements(src)
		if err != nil {
			return nil, err
		}
		if step.Index != n {
			return nil, fmt.Errorf("%w: index %d (length %d)", ErrOutOfRange, step.Index, n)
		}
		container := childContainerType(steps[1])
		inner, err := insertIn(container, steps[1:], rawValue, indent, prefix+indent, colonSpace)
		if err != nil {
			return nil, err
		}
		return insertElementIntoArray(src, step, inner, indent, prefix, colonSpace)
	}
	return nil, fmt.Errorf("expected object or array at path")
}

// childContainerType returns the empty container that step will need
// to descend into.
func childContainerType(step jsonStep) []byte {
	if step.IsIdx || step.IsAppend {
		return []byte("[]")
	}
	return []byte("{}")
}

// addMemberToObject inserts `"key": value` into the object in src. An
// empty object gets the member between {} ; a populated one gets it
// after the last member, with a comma. In pretty mode each new line is
// prefixed with prefix (closing brace) or prefix+indent (member and
// its nested lines). colonSpace picks the colon and comma separators.
func addMemberToObject(src []byte, key string, value []byte, indent, prefix string, colonSpace bool) ([]byte, error) {
	end := bytes.LastIndexByte(src, '}')
	if end < 0 {
		return nil, fmt.Errorf("object has no closing '}'")
	}
	start := bytes.IndexByte(src, '{')
	if start < 0 || start >= end {
		return nil, fmt.Errorf("object has no opening '{'")
	}
	empty := len(bytes.TrimSpace(src[start+1:end])) == 0

	keyJSON, err := json.Marshal(key)
	if err != nil {
		return nil, err
	}

	sep := ":"
	comma := ","
	if colonSpace {
		sep = ": "
		comma = ", "
	}

	var insertion []byte
	var insertAt int
	if empty {
		insertAt = start + 1
		if indent == "" {
			insertion = []byte(string(keyJSON) + sep + string(value))
		} else {
			insertion = []byte("\n" + prefix + indent + string(keyJSON) + sep + string(value) + "\n" + prefix)
		}
	} else {
		var err error
		insertAt, err = lastObjectMemberValueEnd(src)
		if err != nil {
			return nil, err
		}
		if indent == "" {
			insertion = []byte(comma + string(keyJSON) + sep + string(value))
		} else {
			insertion = []byte(",\n" + prefix + indent + string(keyJSON) + sep + string(value))
		}
	}
	return splice(src, insertAt, insertAt, insertion), nil
}

// insertElementIntoArray inserts value into the array in src at the
// position named by step: [N] before element N, [len] or [] at the end.
// colonSpace picks the comma separator in the compact branch.
func insertElementIntoArray(src []byte, step jsonStep, value []byte, indent, prefix string, colonSpace bool) ([]byte, error) {
	n, err := countElements(src)
	if err != nil {
		return nil, err
	}
	start := bytes.IndexByte(src, '[')
	if start < 0 {
		return nil, fmt.Errorf("array has no opening '['")
	}

	comma := ","
	if colonSpace {
		comma = ", "
	}

	if n == 0 {
		if !step.IsAppend && step.Index != 0 {
			return nil, fmt.Errorf("%w: index %d (empty array)", ErrOutOfRange, step.Index)
		}
		var insertion []byte
		if indent == "" {
			insertion = value
		} else {
			insertion = []byte("\n" + prefix + indent + string(value) + "\n" + prefix)
		}
		return splice(src, start+1, start+1, insertion), nil
	}

	if step.IsAppend || step.Index == n {
		_, lastEnd, _ := findElementRange(src, n-1)
		var insertion []byte
		if indent == "" {
			insertion = []byte(comma + string(value))
		} else {
			insertion = []byte(",\n" + prefix + indent + string(value))
		}
		return splice(src, lastEnd, lastEnd, insertion), nil
	}

	if step.Index < 0 || step.Index > n {
		return nil, fmt.Errorf("%w: index %d (length %d)", ErrOutOfRange, step.Index, n)
	}

	elemStart, _, found := findElementRange(src, step.Index)
	if !found {
		return nil, fmt.Errorf("%w: index %d", ErrOutOfRange, step.Index)
	}
	var insertion []byte
	if indent == "" {
		insertion = []byte(string(value) + comma)
	} else {
		insertion = []byte(string(value) + ",\n" + prefix + indent)
	}
	return splice(src, elemStart, elemStart, insertion), nil
}
