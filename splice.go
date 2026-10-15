package jsonedit

// splice replaces src[start:end] with newBytes.
func splice(src []byte, start, end int, newBytes []byte) []byte {
	out := make([]byte, 0, len(src)-(end-start)+len(newBytes))
	out = append(out, src[:start]...)
	out = append(out, newBytes...)
	out = append(out, src[end:]...)
	return out
}

// expandDeleteRange widens [itemStart, itemEnd) to swallow one of the
// surrounding commas, so the result is valid JSON and reads like a
// hand-edit: no trailing comma, no blank line left where the item was.
func expandDeleteRange(src []byte, itemStart, itemEnd int) (int, int) {
	start := itemStart
	end := itemEnd

	// hasFollow walks over all JSON whitespace, not just space and
	// tab. A value followed by a newline and then a comma is valid
	// JSON -- {"a":0\n,"":0} is one example, and the fuzzer found
	// it -- and a space/tab-only skip misses the comma, leaving it
	// behind after the delete.
	p := end
	for p < len(src) && isJSONSpace(src[p]) {
		p++
	}
	hasFollow := p < len(src) && src[p] == ','

	q := start - 1
	for q >= 0 && isJSONSpace(src[q]) {
		q--
	}
	hasPrec := q >= 0 && src[q] == ','

	switch {
	case hasFollow:
		start = lineStartIfOwnLine(src, itemStart)
		end = p + 1
		for end < len(src) && (src[end] == ' ' || src[end] == '\t') {
			end++
		}
	case hasPrec:
		start = q
		end = itemEnd
	default:
		prev := start - 1
		for prev >= 0 && isJSONSpace(src[prev]) {
			prev--
		}
		next := end
		for next < len(src) && isJSONSpace(src[next]) {
			next++
		}
		if prev >= 0 && next < len(src) &&
			(src[prev] == '{' || src[prev] == '[') &&
			(src[next] == '}' || src[next] == ']') {
			start = prev + 1
			end = next
		}
	}
	return start, end
}

// lineStartIfOwnLine returns the position of the line ending that
// precedes the current line when the line consists only of leading
// whitespace followed by the item. If other content shares the line,
// returns itemStart so a preceding "comma space" is not disturbed.
//
// On a CRLF file the function returns the position of the \r, not the
// \n, so the caller removes both bytes of the line ending along with
// the item. Returning the \n alone leaves an orphan \r before the
// next line.
func lineStartIfOwnLine(src []byte, itemStart int) int {
	i := itemStart
	for i > 0 && (src[i-1] == ' ' || src[i-1] == '\t') {
		i--
	}
	if i == 0 || src[i-1] != '\n' {
		return itemStart
	}
	start := i - 1
	if start > 0 && src[start-1] == '\r' {
		start--
	}
	return start
}
