package jsonedit

import "testing"

// isJSONSpace: the four bytes the JSON spec calls whitespace, and
// nothing else.
func TestIsJSONSpace(t *testing.T) {
	for _, c := range []byte{' ', '\t', '\n', '\r'} {
		if !isJSONSpace(c) {
			t.Errorf("isJSONSpace(%q) = false, want true", c)
		}
	}
	for _, c := range []byte{'a', '0', '"', ',', ':', '{', 0, 0xff} {
		if isJSONSpace(c) {
			t.Errorf("isJSONSpace(%q) = true, want false", c)
		}
	}
}

// findKeyStart walks back from a value's start to the opening quote
// of its key. It is the anchor for Delete on an object member.
func TestFindKeyStart(t *testing.T) {
	cases := []struct {
		name  string
		src   string
		value int
		want  int
	}{
		{"simple", `{"a": 1}`, 6, 1},
		{"compact", `{"a":1}`, 5, 1},
		{"space before colon", `{"a" : 1}`, 7, 1},
		{"second key", `{"a": 1, "b": 2}`, 14, 9},
		{"escaped quote in key", `{"a\"b": 1}`, 9, 1},
		{"backslash in key", `{"a\\b": 1}`, 9, 1},
		{"empty key", `{"": 1}`, 5, 1},
		{"array element has no key", `[1, 2]`, 1, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := findKeyStart([]byte(tc.src), tc.value)
			if got != tc.want {
				t.Fatalf("findKeyStart(%q, %d) = %d, want %d",
					tc.src, tc.value, got, tc.want)
			}
		})
	}
}

// findMemberValueRange returns the [start, end) of the value for a
// key inside an object.
func TestFindMemberValueRange(t *testing.T) {
	cases := []struct {
		name      string
		src       string
		key       string
		vs, ve    int
		wantFound bool
	}{
		{"empty object", `{}`, "a", 0, 0, false},
		{"simple", `{"a": 1}`, "a", 6, 7, true},
		{"compact", `{"a":1}`, "a", 5, 6, true},
		{"second key", `{"a": 1, "b": 2}`, "b", 14, 15, true},
		{"missing key", `{"a": 1}`, "b", 0, 0, false},
		{"nested object value", `{"a": {"b": 1}}`, "a", 6, 14, true},
		{"value is array", `{"a": [1, 2]}`, "a", 6, 12, true},
		{"escaped quote in key", `{"a\"b": 1}`, `a"b`, 9, 10, true},
		{"empty key", `{"": 1}`, "", 5, 6, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vs, ve, found := findMemberValueRange([]byte(tc.src), tc.key)
			if found != tc.wantFound {
				t.Fatalf("findMemberValueRange(%q, %q) found=%v, want %v",
					tc.src, tc.key, found, tc.wantFound)
			}
			if found && (vs != tc.vs || ve != tc.ve) {
				t.Fatalf("findMemberValueRange(%q, %q) = [%d, %d), want [%d, %d)",
					tc.src, tc.key, vs, ve, tc.vs, tc.ve)
			}
		})
	}
}

// findElementRange returns the [start, end) of the element at index
// in an array.
func TestFindElementRange(t *testing.T) {
	cases := []struct {
		name      string
		src       string
		index     int
		es, ee    int
		wantFound bool
	}{
		{"empty array", `[]`, 0, 0, 0, false},
		{"single", `[1]`, 0, 1, 2, true},
		{"middle of spaced", `[1, 2, 3]`, 1, 4, 5, true},
		{"middle of compact", `[1,2,3]`, 1, 3, 4, true},
		{"tail", `[1, 2, 3]`, 2, 7, 8, true},
		{"out of range", `[1, 2, 3]`, 3, 0, 0, false},
		{"not array", `{"a": 1}`, 0, 0, 0, false},
		{"nested value", `[{"a": 1}]`, 0, 1, 9, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			es, ee, found := findElementRange([]byte(tc.src), tc.index)
			if found != tc.wantFound {
				t.Fatalf("findElementRange(%q, %d) found=%v, want %v",
					tc.src, tc.index, found, tc.wantFound)
			}
			if found && (es != tc.es || ee != tc.ee) {
				t.Fatalf("findElementRange(%q, %d) = [%d, %d), want [%d, %d)",
					tc.src, tc.index, es, ee, tc.es, tc.ee)
			}
		})
	}
}

// countElements returns the element count of an array; non-arrays
// and unclosed arrays are errors.
func TestCountElements(t *testing.T) {
	cases := []struct {
		name    string
		src     string
		want    int
		wantErr bool
	}{
		{"empty", `[]`, 0, false},
		{"one", `[1]`, 1, false},
		{"three spaced", `[1, 2, 3]`, 3, false},
		{"three compact", `[1,2,3]`, 3, false},
		{"nested", `[{"a": 1}, {"b": 2}]`, 2, false},
		{"with whitespace", "[\n  1,\n  2\n]", 2, false},
		{"empty with whitespace", "[\n]", 0, false},
		{"not array", `{"a": 1}`, 0, true},
		{"empty string", ``, 0, true},
		{"unclosed", `[1, 2`, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := countElements([]byte(tc.src))
			if (err != nil) != tc.wantErr {
				t.Fatalf("countElements(%q) err=%v, wantErr=%v",
					tc.src, err, tc.wantErr)
			}
			if !tc.wantErr && got != tc.want {
				t.Fatalf("countElements(%q) = %d, want %d", tc.src, got, tc.want)
			}
		})
	}
}

// lastObjectMemberValueEnd is the byte position right after the last
// member's value. Insert uses it as the anchor for a new member.
func TestLastObjectMemberValueEnd(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want int
	}{
		{"empty object", `{}`, 1},
		{"single member", `{"a": 1}`, 7},
		{"two members", `{"a": 1, "b": 2}`, 15},
		{"nested value", `{"a": {"x": 1}}`, 14},
		{"spaced value", `{"a" : 1 }`, 8},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := lastObjectMemberValueEnd([]byte(tc.src))
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("lastObjectMemberValueEnd(%q) = %d, want %d",
					tc.src, got, tc.want)
			}
		})
	}
}

// locateParentRange returns the byte range of the container that
// holds the last step of steps. With one step or none, the container
// is the root.
func TestLocateParentRange(t *testing.T) {
	src := []byte(`{"a": 1}`)
	s, e, err := locateParentRange(src, []jsonStep{{Key: "a"}})
	if err != nil {
		t.Fatal(err)
	}
	if s != 0 || e != len(src) {
		t.Fatalf("one step must give [0, %d), got [%d, %d)", len(src), s, e)
	}

	src = []byte(`{"a": {"b": 1}}`)
	s, e, err = locateParentRange(src, []jsonStep{{Key: "a"}, {Key: "b"}})
	if err != nil {
		t.Fatal(err)
	}
	if string(src[s:e]) != `{"b": 1}` {
		t.Fatalf("parent range = %q, want %q", src[s:e], `{"b": 1}`)
	}

	// Zero steps also gives the root.
	src = []byte(`[1, 2]`)
	s, e, err = locateParentRange(src, nil)
	if err != nil {
		t.Fatal(err)
	}
	if s != 0 || e != len(src) {
		t.Fatalf("zero steps must give [0, %d), got [%d, %d)", len(src), s, e)
	}
}
