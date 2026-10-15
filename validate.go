package jsonedit

import (
	"encoding/json"
	"fmt"
)

// Validate reports whether src is valid JSON. It is the read-only
// twin of Set, Delete, Insert, Append, and Rename: a document that
// Validate accepts is one the editors can edit, and a document it
// rejects fails with a byte offset that names where.
//
// json.Valid scans the whole input without building a tree; it is
// the fast path. On failure, Unmarshal runs once to produce the
// error with an offset — the input is already known bad, so the
// cost of building a tree is paid only on the failure path.
// Validate reports whether src is valid JSON. It is the read-only
// twin of Set, Delete, Insert, Append, and Rename: a document that
// Validate accepts is one the editors can edit, and a document it
// rejects fails with a byte offset that names where.
//
// json.Valid scans the whole input without building a tree; it is
// the fast path. On failure, Unmarshal runs once to produce the
// error with an offset — the input is already known bad, so the
// cost of building a tree is paid only on the failure path.
//
// The error wraps ErrInvalidSource.
func Validate(src []byte) error {
	if json.Valid(src) {
		return nil
	}
	var v any
	if err := json.Unmarshal(src, &v); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidSource, err)
	}
	// json.Valid said no, Unmarshal said yes. Not expected, but if
	// it happens the two disagree; the honest answer is still
	// "invalid source".
	return fmt.Errorf("%w: json.Valid and json.Unmarshal disagree", ErrInvalidSource)
}
