// Package jsonedit edits JSON documents by path.
//
// The package operates by byte-range splice: it locates the range
// of the value at a JSON path in the source and replaces exactly
// that range. The document is never round-tripped through
// json.Unmarshal and json.Marshal, because that would reformat the
// whole file and produce a diff far larger than the change. Every
// byte outside the target range is preserved.
//
// "Value" is the JSON spec's term (RFC 8259): an object, an array,
// a number, a string, true, false, or null. A path names one of
// them. It is not validation: Validate and Locate are read-only
// supporting functions, not the package's purpose.
//
// Input must be valid JSON; a successful edit returns valid JSON.
// Whitespace outside the affected byte range is preserved
// byte-for-byte; the affected range may be reserialized. A new
// object member emitted by Insert, Append, or Rename follows the
// file's own colon and comma spacing: `": "` and `", "` when the
// first colon outside a string is followed by a space, `":"` and
// `","` otherwise. A document with no colon at all -- an array, an
// empty object -- uses the spaced form. Set emits primitives
// compactly and objects and arrays with the file's detected
// indentation.
//
// The public API is Set, Delete, Insert, Append, Rename, Locate,
// and Validate. A future package for another structured format
// (xmlval, yamlval) may follow the same shape.
//
// Every error wraps a sentinel. Use errors.Is to discriminate:
// ErrInvalidSource, ErrInvalidValue, ErrInvalidPath, ErrNotFound,
// ErrOutOfRange, ErrTypeMismatch, ErrCollision, ErrExpectMismatch,
// ErrRootOp. The message carries the detail.
package jsonedit
