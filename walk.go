package jsonedit

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// isJSONSpace reports whether c is one of the four bytes the JSON spec
// treats as whitespace.
func isJSONSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// locateJSONItem returns the byte range [start, end) of the value at
// path in src, plus a flag saying whether the value is a member of an
// object (as opposed to an element of an array). The flag lets delete
// widen the range to include the key.
//
// Positions come from json.Decoder.InputOffset, which reports the byte
// offset of the end of the most recently read token; the start of a
// token is found by skipping whitespace, commas, and colons forward
// from the previous token's end.
func locateJSONItem(src []byte, path string) (int, int, bool, error) {
	steps, err := parseJSONPath(path)
	if err != nil {
		return 0, 0, false, err
	}
	dec := json.NewDecoder(bytes.NewReader(src))
	dec.UseNumber()
	last := 0
	return walkJSON(src, dec, &last, steps)
}

func walkJSON(src []byte, dec *json.Decoder, last *int, path []jsonStep) (int, int, bool, error) {
	start := *last
	for start < len(src) && (isJSONSpace(src[start]) || src[start] == ':' || src[start] == ',') {
		start++
	}
	tok, err := dec.Token()
	if err != nil {
		return 0, 0, false, err
	}
	end := int(dec.InputOffset())
	*last = end

	delim, isDelim := tok.(json.Delim)
	if !isDelim {
		if len(path) != 0 {
			return 0, 0, false, fmt.Errorf("%w: value is a primitive at %q", ErrTypeMismatch, path[0].Key)
		}
		return start, end, false, nil
	}

	switch delim {
	case '{':
		if len(path) == 0 {
			if err := skipJSONObject(dec, last); err != nil {
				return 0, 0, false, err
			}
			return start, *last, false, nil
		}
		step := path[0]
		if step.IsIdx {
			return 0, 0, false, fmt.Errorf("%w: expected object key, got [%d]", ErrTypeMismatch, step.Index)
		}
		for dec.More() {
			keyTok, err := dec.Token()
			if err != nil {
				return 0, 0, false, err
			}
			key, _ := keyTok.(string)
			*last = int(dec.InputOffset())
			if key == step.Key {
				if len(path) == 1 {
					vs, ve, _, err := walkJSON(src, dec, last, nil)
					if err != nil {
						return 0, 0, false, err
					}
					return vs, ve, true, nil
				}
				return walkJSON(src, dec, last, path[1:])
			}
			if err := skipJSONValue(dec, last); err != nil {
				return 0, 0, false, err
			}
		}
		if _, err := dec.Token(); err != nil {
			return 0, 0, false, err
		}
		*last = int(dec.InputOffset())
		return 0, 0, false, fmt.Errorf("%w: key %q", ErrNotFound, step.Key)
	case '[':
		if len(path) == 0 {
			if err := skipJSONArray(dec, last); err != nil {
				return 0, 0, false, err
			}
			return start, *last, false, nil
		}
		step := path[0]
		if !step.IsIdx {
			return 0, 0, false, fmt.Errorf("%w: expected array index, got key %q", ErrTypeMismatch, step.Key)
		}
		for i := 0; dec.More(); i++ {
			if i == step.Index {
				return walkJSON(src, dec, last, path[1:])
			}
			if err := skipJSONValue(dec, last); err != nil {
				return 0, 0, false, err
			}
		}
		if _, err := dec.Token(); err != nil {
			return 0, 0, false, err
		}
		*last = int(dec.InputOffset())
		return 0, 0, false, fmt.Errorf("%w: index %d", ErrOutOfRange, step.Index)
	}
	return 0, 0, false, fmt.Errorf("unexpected delimiter %q", delim)
}

// findKeyStart walks back from a value's start to the opening quote of
// its key. Returns valueStart when no key is found, so an array
// element is treated as its own item.
func findKeyStart(src []byte, valueStart int) int {
	i := valueStart - 1
	for i >= 0 && (isJSONSpace(src[i]) || src[i] == ':') {
		i--
	}
	if i < 0 || src[i] != '"' {
		return valueStart
	}
	j := i - 1
	for j >= 0 {
		if src[j] == '"' {
			backslashes := 0
			for k := j - 1; k >= 0 && src[k] == '\\'; k-- {
				backslashes++
			}
			if backslashes%2 == 0 {
				return j
			}
		}
		j--
	}
	return valueStart
}

func skipJSONValue(dec *json.Decoder, last *int) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	*last = int(dec.InputOffset())
	delim, ok := tok.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		return skipJSONObject(dec, last)
	case '[':
		return skipJSONArray(dec, last)
	}
	return nil
}

func skipJSONObject(dec *json.Decoder, last *int) error {
	for dec.More() {
		if _, err := dec.Token(); err != nil {
			return err
		}
		*last = int(dec.InputOffset())
		if err := skipJSONValue(dec, last); err != nil {
			return err
		}
	}
	if _, err := dec.Token(); err != nil {
		return err
	}
	*last = int(dec.InputOffset())
	return nil
}

func skipJSONArray(dec *json.Decoder, last *int) error {
	for dec.More() {
		if err := skipJSONValue(dec, last); err != nil {
			return err
		}
	}
	if _, err := dec.Token(); err != nil {
		return err
	}
	*last = int(dec.InputOffset())
	return nil
}

// skipJSONValueAt returns the byte position right after the JSON value
// that starts at src[i]. It creates a fresh decoder on the tail so the
// caller can advance through an array or object without sharing state.
func skipJSONValueAt(src []byte, i int) (int, error) {
	dec := json.NewDecoder(bytes.NewReader(src[i:]))
	dec.UseNumber()
	var last int
	if err := skipJSONValue(dec, &last); err != nil {
		return 0, err
	}
	return i + last, nil
}

// findMemberValueRange returns the byte range [start, end) of the value
// for key inside the object in src. found is false when the key is
// absent.
func findMemberValueRange(src []byte, key string) (int, int, bool) {
	dec := json.NewDecoder(bytes.NewReader(src))
	dec.UseNumber()
	tok, err := dec.Token()
	if err != nil {
		return 0, 0, false
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return 0, 0, false
	}
	last := int(dec.InputOffset())
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return 0, 0, false
		}
		k, _ := keyTok.(string)
		last = int(dec.InputOffset())
		p := last
		for p < len(src) && (isJSONSpace(src[p]) || src[p] == ':') {
			p++
		}
		valStart := p
		if err := skipJSONValue(dec, &last); err != nil {
			return 0, 0, false
		}
		if k == key {
			return valStart, last, true
		}
	}
	return 0, 0, false
}

// findElementRange returns the byte range [start, end) of the element
// at index in the array in src. found is false when the index is out
// of range or src is not an array.
func findElementRange(src []byte, index int) (int, int, bool) {
	i := 0
	for i < len(src) && isJSONSpace(src[i]) {
		i++
	}
	if i >= len(src) || src[i] != '[' {
		return 0, 0, false
	}
	i++
	for n := 0; ; n++ {
		for i < len(src) && isJSONSpace(src[i]) {
			i++
		}
		if i >= len(src) || src[i] == ']' {
			return 0, 0, false
		}
		elemStart := i
		j, err := skipJSONValueAt(src, i)
		if err != nil {
			return 0, 0, false
		}
		if n == index {
			return elemStart, j, true
		}
		i = j
		for i < len(src) && isJSONSpace(src[i]) {
			i++
		}
		if i < len(src) && src[i] == ',' {
			i++
			continue
		}
		return 0, 0, false
	}
}

// countElements returns the number of elements in the array in src.
func countElements(src []byte) (int, error) {
	i := 0
	for i < len(src) && isJSONSpace(src[i]) {
		i++
	}
	if i >= len(src) || src[i] != '[' {
		return 0, fmt.Errorf("not an array")
	}
	i++
	n := 0
	for {
		for i < len(src) && isJSONSpace(src[i]) {
			i++
		}
		if i >= len(src) {
			return 0, fmt.Errorf("unexpected end of array")
		}
		if src[i] == ']' {
			return n, nil
		}
		n++
		j, err := skipJSONValueAt(src, i)
		if err != nil {
			return 0, err
		}
		i = j
		for i < len(src) && isJSONSpace(src[i]) {
			i++
		}
		if i < len(src) && src[i] == ',' {
			i++
			continue
		}
		if i < len(src) && src[i] == ']' {
			return n, nil
		}
		return 0, fmt.Errorf("unexpected character at %d", i)
	}
}

// lastObjectMemberValueEnd returns the byte position right after the
// last member's value in the object in src. Inserting a comma there,
// rather than before the closing brace, keeps the line break that
// precedes the brace.
//
// The decoder uses UseNumber so a number of any length is read as a
// json.Number, not converted to float64. Without it, a number whose
// value overflows float64 makes Token fail, and the function would
// return the position right after the last key -- splicing the new
// member into the middle of the previous pair.
func lastObjectMemberValueEnd(src []byte) (int, error) {
	dec := json.NewDecoder(bytes.NewReader(src))
	dec.UseNumber()
	last := 0
	if _, err := dec.Token(); err != nil {
		return 0, err
	}
	last = int(dec.InputOffset())
	for dec.More() {
		if _, err := dec.Token(); err != nil {
			return 0, err
		}
		last = int(dec.InputOffset())
		if err := skipJSONValue(dec, &last); err != nil {
			return 0, err
		}
	}
	return last, nil
}

// locateParentRange returns the byte range [start, end) of the
// container that holds the last step of steps. When steps has one
// element, the container is the root, and the range is the whole
// source.
func locateParentRange(src []byte, steps []jsonStep) (int, int, error) {
	if len(steps) <= 1 {
		return 0, len(src), nil
	}
	parentSteps := steps[:len(steps)-1]
	dec := json.NewDecoder(bytes.NewReader(src))
	dec.UseNumber()
	last := 0
	s, e, _, err := walkJSON(src, dec, &last, parentSteps)
	if err != nil {
		return 0, 0, err
	}
	return s, e, nil
}
