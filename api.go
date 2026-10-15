package jsonedit

import (
	"encoding/json"
	"fmt"
)

// Locate returns the byte range of the value at path in src. The
// source must be valid JSON. Locate is a read-only inspection: it does
// not modify src.
func Locate(src []byte, path string) (Range, error) {
	if !json.Valid(src) {
		return Range{}, ErrInvalidSource
	}
	s, e, _, err := locateJSONItem(src, path)
	if err != nil {
		return Range{}, err
	}
	return Range{Start: s, End: e}, nil
}

// Set replaces the value at path with rawValue. rawValue must be a
// valid JSON literal. When hasExpect is true, the current value at
// path must be semantically equal to expect or the edit is refused
// and src is unchanged.
//
// A new object or array value is pretty-printed with the file's
// detected indent and the current line's prefix; a primitive is
// emitted compactly. src must be valid JSON.
func Set(src []byte, path, rawValue, expect string, hasExpect bool) ([]byte, error) {
	if !json.Valid(src) {
		return nil, ErrInvalidSource
	}
	if !json.Valid([]byte(rawValue)) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidValue, rawValue)
	}
	s, e, _, err := locateJSONItem(src, path)
	if err != nil {
		return nil, err
	}
	if hasExpect {
		if err := assertJSONValue(src[s:e], expect); err != nil {
			return nil, err
		}
	}
	indent := detectJSONIndent(src)
	prefix := lineIndentPrefix(src, s)
	newBytes, err := serializeJSONValue(rawValue, indent, prefix)
	if err != nil {
		return nil, err
	}
	return splice(src, s, e, newBytes), nil
}

// Delete removes the value at path. When hasExpect is true, the
// current value at path must be semantically equal to expect or the
// edit is refused. The root cannot be deleted; use a full-file write
// instead.
//
// The byte range is widened to swallow one of the surrounding commas
// and, for an item on its own line, the preceding line ending. The
// result is valid JSON and reads like a hand-edit: no blank line, no
// trailing comma.
func Delete(src []byte, path, expect string, hasExpect bool) ([]byte, error) {
	if !json.Valid(src) {
		return nil, ErrInvalidSource
	}
	steps, err := parseJSONPath(path)
	if err != nil {
		return nil, err
	}
	if len(steps) == 0 {
		return nil, fmt.Errorf("cannot delete the root: %w", ErrRootOp)
	}
	vs, ve, inObj, err := locateJSONItem(src, path)
	if err != nil {
		return nil, err
	}
	if hasExpect {
		if err := assertJSONValue(src[vs:ve], expect); err != nil {
			return nil, err
		}
	}
	itemStart := vs
	if inObj {
		itemStart = findKeyStart(src, vs)
	}
	s, e := expandDeleteRange(src, itemStart, ve)
	return splice(src, s, e, nil), nil
}

// Insert adds a new value at path. The last step of the path chooses
// the shape:
//
//   - a key:   a new member of the object at steps[:len-1]. A key
//     that already exists is an error; use Set.
//   - [N]:     a new element of the array at steps[:len-1], position
//     N. Any N in [0, len] is valid: [len] appends.
//   - []:      a new element appended to the array at steps[:len-1].
//
// Containers along the path are auto-created if missing. Insert never
// replaces.
func Insert(src []byte, path, rawValue string) ([]byte, error) {
	if !json.Valid(src) {
		return nil, ErrInvalidSource
	}
	if !json.Valid([]byte(rawValue)) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidValue, rawValue)
	}
	steps, err := parseJSONPath(path)
	if err != nil {
		return nil, err
	}
	if len(steps) == 0 {
		return nil, fmt.Errorf("cannot insert at root: %w", ErrRootOp)
	}
	indent := detectJSONIndent(src)
	colonSpace := detectColonSpace(src)
	return insertIn(src, steps, rawValue, indent, "", colonSpace)
}

// Append adds rawValue to the array at path. The path must end in [];
// the value is appended at the array's end, and the container is
// auto-created if missing.
func Append(src []byte, path, rawValue string) ([]byte, error) {
	if !json.Valid(src) {
		return nil, ErrInvalidSource
	}
	if !json.Valid([]byte(rawValue)) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidValue, rawValue)
	}
	steps, err := parseJSONPath(path)
	if err != nil {
		return nil, err
	}
	if len(steps) == 0 || !steps[len(steps)-1].IsAppend {
		return nil, fmt.Errorf("append requires a path ending in []: %w", ErrInvalidPath)
	}
	indent := detectJSONIndent(src)
	colonSpace := detectColonSpace(src)
	return insertIn(src, steps, rawValue, indent, "", colonSpace)
}

// Rename changes the key of the member at path. The path names an
// existing object key; newKey is the raw new name, not a JSON literal.
// The value's bytes are reused unchanged.
//
// A new name that already exists in the same object is an error; two
// members with the same name are technically valid JSON but almost
// always a bug. When hasExpect is true, the current value must be
// semantically equal to expect.
func Rename(src []byte, path, newKey, expect string, hasExpect bool) ([]byte, error) {
	if !json.Valid(src) {
		return nil, ErrInvalidSource
	}
	steps, err := parseJSONPath(path)
	if err != nil {
		return nil, err
	}
	if len(steps) == 0 {
		return nil, fmt.Errorf("cannot rename the root: %w", ErrRootOp)
	}
	last := steps[len(steps)-1]
	if last.IsIdx || last.IsAppend {
		return nil, fmt.Errorf("rename only works on object keys: %w", ErrTypeMismatch)
	}
	vs, ve, inObj, err := locateJSONItem(src, path)
	if err != nil {
		return nil, err
	}
	if !inObj {
		return nil, fmt.Errorf("rename only works on object keys: %w", ErrTypeMismatch)
	}
	if hasExpect {
		if err := assertJSONValue(src[vs:ve], expect); err != nil {
			return nil, err
		}
	}
	keyStart := findKeyStart(src, vs)
	if keyStart == vs {
		return nil, fmt.Errorf("could not locate the key")
	}
	ps, pe, err := locateParentRange(src, steps)
	if err != nil {
		return nil, err
	}
	if _, _, found := findMemberValueRange(src[ps:pe], newKey); found {
		return nil, fmt.Errorf("%w: %q", ErrCollision, newKey)
	}
	keyJSON, err := json.Marshal(newKey)
	if err != nil {
		return nil, err
	}
	sep := ":"
	if detectColonSpace(src) {
		sep = ": "
	}
	newPair := make([]byte, 0, len(keyJSON)+len(sep)+(ve-vs))
	newPair = append(newPair, keyJSON...)
	newPair = append(newPair, sep...)
	newPair = append(newPair, src[vs:ve]...)
	out := make([]byte, 0, len(src)-(ve-keyStart)+len(newPair))
	out = append(out, src[:keyStart]...)
	out = append(out, newPair...)
	out = append(out, src[ve:]...)
	return out, nil
}
