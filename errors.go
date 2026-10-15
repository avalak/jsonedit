package jsonedit

import "errors"

// Every error returned by the public API wraps one of these
// sentinels. errors.Is discriminates the category; the message
// carries the detail (which key, which index, what was expected).
//
// Internal-only errors -- invariant violations a caller with valid
// input cannot trigger -- are plain fmt.Errorf and wrap nothing.
var (
	// ErrInvalidSource is returned when src is not valid JSON.
	ErrInvalidSource = errors.New("source is not valid JSON")

	// ErrInvalidValue is returned when rawValue or expect is not a
	// valid JSON literal. The empty string is not a valid value.
	ErrInvalidValue = errors.New("value is not valid JSON")

	// ErrInvalidPath is returned when a path is malformed, or when
	// a valid path cannot be used for the requested operation: an
	// Append on a path that does not end in [].
	ErrInvalidPath = errors.New("invalid path")

	// ErrNotFound is returned when a path names an object key that
	// does not exist.
	ErrNotFound = errors.New("not found")

	// ErrOutOfRange is returned when a path names an array index
	// outside [0, len]. For Insert, [len] is valid: it appends.
	ErrOutOfRange = errors.New("out of range")

	// ErrTypeMismatch is returned when a path step walks through a
	// value of the wrong type: a key into an array, an index into
	// an object, or any step into a primitive.
	ErrTypeMismatch = errors.New("type mismatch")

	// ErrCollision is returned when Insert or Rename would create
	// a duplicate key in the same object.
	ErrCollision = errors.New("key already exists")

	// ErrExpectMismatch is returned when hasExpect is true and the
	// value at path does not equal expect. The comparison is
	// semantic: key order, whitespace, and 1 versus 1.0 do not
	// matter.
	ErrExpectMismatch = errors.New("expect-value mismatch")

	// ErrRootOp is returned when an operation is not allowed on
	// the document root: Delete("$") and Rename("$").
	ErrRootOp = errors.New("operation not allowed on the root")
)
