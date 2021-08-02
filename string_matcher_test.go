package hex

import (
	"fmt"
	"testing"
)

// TODO: Test string matcher

func TestStringMatcher(t *testing.T) {
	testCases := []struct {
		arg   interface{}
		input string
		want  bool
	}{
		// Exact matching strings
		{"foo", "foo", true},
		{"foo", "bar", false},
		{"", "bar", false},

		// Regexp matching
		{R("foo"), "barfoobar", true},
		{R("^foo"), "foobar", true},
		{R("foo$"), "barfoo", true},
		{R(`^\d+$`), "123", true},
		{R(`^\d+$`), "123.0", false},

		// Any/None
		{Any, "foo", true},
		{Any, "", true},
		{Any, "1023", true},
		{None, "foo", false},
		{None, "", false},
		{None, "1023", false},

		// Custom matching funcs
		{func(s string) bool { return s == "foo" }, "foo", true},
		{func(s string) bool { return s == "foo" }, "bar", false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Matching %v against %q", tc.arg, tc.input), func(t *testing.T) {
			matcher, err := makeStringMatcher(tc.arg)
			if err != nil {
				t.Errorf("Unexpected error creating matcher for argument %v", tc.arg)
			}
			got := matcher.match(tc.input)

			if got != tc.want {
				t.Errorf("Failed: Got %t, want %t", got, tc.want)
				t.FailNow()
			}
		})
	}
}

func TestStringMatcherInvalidArgs(t *testing.T) {
	if _, err := makeStringMatcher(struct{}{}); err == nil {
		t.Errorf("Expected makeStringMatcher with invalid arguments to return an error, but non was returned")
	}
}
