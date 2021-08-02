package hex

import (
	"fmt"
	"net/url"
)

type keyValueMatcher struct {
	key   stringMatcher
	value stringMatcher
}

type urlValuesMatcher struct {
	pairs []keyValueMatcher
}

func (u *urlValuesMatcher) matches(values url.Values) bool {
	for _, pair := range u.pairs {
		for key, values := range values {
			if pair.key.match(key) {
				for _, value := range values {
					if pair.value.match(value) {
						return true
					}
				}
			}
		}
	}
	return false
}

func makeValueMatcher(arg interface{}) (urlValuesMatcher, error) {
	// Single argument. If it's a string/regexp/stringmatcher, then we treat it like a key/value matcher,
	// with "Any" as the value.
	if m, err := makeStringMatcher(arg); err == nil {
		return urlValuesMatcher{
			pairs: []keyValueMatcher{
				{key: m, value: mustMakeStringMatcher(Any)},
			},
		}, nil
	} else if params, ok := arg.(P); ok {
		pairs := make([]keyValueMatcher, 0, len(params))
		for key, value := range params {
			pairs = append(pairs, keyValueMatcher{
				key:   mustMakeStringMatcher(key),
				value: mustMakeStringMatcher(value),
			})
		}
		return urlValuesMatcher{pairs: pairs}, nil
	}

	return urlValuesMatcher{}, fmt.Errorf("Cannot use value %v when matching against url.Values", arg)
}

func makeURLValuesMatcher(args []interface{}) (urlValuesMatcher, error) {
	if len(args) == 0 {
		// return a matcher that matches any non-empty url.Values{}
		return urlValuesMatcher{
			pairs: []keyValueMatcher{
				{key: mustMakeStringMatcher(Any), value: mustMakeStringMatcher(Any)},
			},
		}, nil
	} else if len(args) == 1 {
		return makeValueMatcher(args[0])
	} else if len(args) == 2 {
		// Double argument. It's a key/value pair, which is essentially a single element map
		return makeURLValuesMatcher([]interface{}{P{args[0]: args[1]}})
	}

	return urlValuesMatcher{}, fmt.Errorf("Cannot use value %v when matching against url.Values", args)
}

func (u *urlValuesMatcher) match(value interface{}) bool {
	return false
}
