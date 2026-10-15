package jsonedit

// Range is the half-open byte interval [Start, End) of a value in the
// source document. Returned by Locate.
type Range struct {
	Start, End int
}
