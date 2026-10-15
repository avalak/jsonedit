package jsonedit

import (
	"strings"
	"testing"
)

func TestValidateOK(t *testing.T) {
	for _, src := range []string{
		`{}`,
		`[]`,
		`{"a": 1}`,
		`[1, 2, 3]`,
		`null`,
		`true`,
		`false`,
		`"string"`,
		`  42  `,
		"{\n  \"a\": [1, 2, 3]\n}\n",
		`{"ключ": "значение"}`,
		`{"emoji": "😀"}`,
	} {
		if err := Validate([]byte(src)); err != nil {
			t.Errorf("Validate(%q) = %v, want nil", src, err)
		}
	}
}

func TestValidateBad(t *testing.T) {
	cases := []string{
		``,
		`{`,
		`[`,
		`{"a": }`,
		`{"a": 1,}`,
		`{"a": 1} trailing`,
		`// comment`,
		`{'single': 1}`,
		`undefined`,
	}
	for _, src := range cases {
		err := Validate([]byte(src))
		if err == nil {
			t.Errorf("Validate(%q) accepted, want error", src)
			continue
		}
		if strings.TrimSpace(err.Error()) == "" {
			t.Errorf("Validate(%q) returned an empty error", src)
		}
	}
}
