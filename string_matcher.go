package hex

import (
	"fmt"
	"regexp"
)

// StringMatcher is used for matching parts of a request that can only ever be strings, such as the HTTP Method,
// path, query string/header/form keys (not values, which can be arrays), etc.
type StringMatcher func(string) bool

// makeStringMatcher takes as input a string, regular expression, or StringMatcher and returns a StringMatcher
func makeStringMatcher(arg interface{}) (stringMatcher, error) {
	if str, ok := arg.(string); ok {
		return &stringLiteralMatcher{str: str}, nil
	} else if re, ok := arg.(*regexp.Regexp); ok {
		return &stringRegexMatcher{pattern: re}, nil
	} else if sm, ok := arg.(func(string) bool); ok {
		return &funcStringMatcher{fn: sm}, nil
	} else if c, ok := arg.(MatchConst); ok {
		if c == Any {
			return &funcStringMatcher{
				fn: func(string) bool { return true },
			}, nil
		} else if c == None {
			return &funcStringMatcher{
				fn: func(string) bool { return false },
			}, nil
		}
	}

	return nil, fmt.Errorf("Cannot use value %v when matching against strings", arg)
}

func mustMakeStringMatcher(value interface{}) stringMatcher {
	if m, err := makeStringMatcher(value); err != nil {
		panic(err)
	} else {
		return m
	}
}

// MatchConst is used to define some built-in matchers with predefined
// behavior, namely All or None
type MatchConst int

const (
	// Any matches anything, matching against Any will always return true
	Any MatchConst = 1

	// None matches nothing, matching against None will always return false
	None MatchConst = 2
)

type stringMatcher interface {
	match(string) bool
	String() string
}

type stringLiteralMatcher struct {
	str string
}

var _ stringMatcher = &stringLiteralMatcher{}

func (s *stringLiteralMatcher) String() string {
	return s.str
}

func (s *stringLiteralMatcher) match(candidate string) bool {
	return s.str == candidate
}

type stringRegexMatcher struct {
	pattern *regexp.Regexp
}

var _ stringMatcher = &stringRegexMatcher{}

func (s *stringRegexMatcher) String() string {
	return s.pattern.String()
}

func (s *stringRegexMatcher) match(candidate string) bool {
	return s.pattern.MatchString(candidate)
}

type noStringMatcher struct{}

var _ stringMatcher = &noStringMatcher{}

func (*noStringMatcher) match(string) bool {
	return false
}

func (*noStringMatcher) String() string {
	return ""
}

type funcStringMatcher struct {
	fn StringMatcher
}

var _ stringMatcher = &funcStringMatcher{}

func (s *funcStringMatcher) match(candidate string) bool {
	return s.fn(candidate)
}

func (s *funcStringMatcher) String() string {
	return "custom string matching function"
}
