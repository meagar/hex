package hex

import (
	"net/url"
	"testing"
)

func TestMatcher(t *testing.T) {
	// empty := url.Values{}
	// single := url.Values{"name": []string{"bob"}}
	// multi := url.Values{"name": []string{"bob"}, "age": []string{"45"}, "email": []string{"bob@example.com"}}

	type Args []interface{}
	type Values url.Values

	testCases := []struct {
		args          Args
		encodedParams string
		want          bool
	}{
		// A completely empty args just tests for a non-empty Values
		{Args{}, "", false},           // Mismatch, no Values
		{Args{}, "key", true},         // Match, Values is non-empty
		{Args{Any, Any}, "", false},   // Mismatch, only matches any non-empty set
		{Args{Any, "foo"}, "", false}, // Mismatch, any key matching a specific value vs an empty set

		// single values that match any parameter name
		{Args{"key"}, "key", true},              // Match, key present but no value
		{Args{"key"}, "key=value", true},        // Match, value doesn't matter
		{Args{"key"}, "", false},                // Mismatch, no query string
		{Args{"key"}, "key1=value", false},      // Mismatch, key not found
		{Args{"key", Any}, "key1=value", false}, // Mismatch, key not found

		// single regex values that match any parameter
		{Args{R(`^key\d+$`)}, "key1", true},
		{Args{R(`^key\d+$`)}, "key1=value1", true},
		{Args{R(`^key\d+$`)}, "skey1=value1", false},

		// simple key/value pairs
		{Args{"key", "value"}, "key=value", true},   // Match
		{Args{"key", "value"}, "", false},           // Mismatch no query string
		{Args{"key", "value"}, "key1=value", false}, // Mismatch because key is not present
		{Args{"key", "value"}, "key=value2", false}, // Mismatch because value is wrong

		// key/value pairs with regexp value
		{Args{"key", R(`value\d`)}, "key=value1", true},      // Match
		{Args{"key", R(`value\d$`)}, "key=value1foo", false}, // No match, anchor not satisfied
		{Args{Any, "value"}, "foo=value", true},              // Match any key, with the given value

		// key/value pairs with regexp key and value
		{Args{R(`^key\d+$`), R(`^value\d+$`)}, "key12=value34", true}, // Match
		{Args{R(`^key\d+$`), R(`value\d$`)}, "key=value1foo", false},  // No match, anchor not satisfied

		{Args{P{"key1": "value1", "key2": "value2"}}, "key1=value1&key2=value2", true},
	}

	for _, tc := range testCases {
		values, err := url.ParseQuery(tc.encodedParams)
		if err != nil {
			panic("Invalid query string:" + tc.encodedParams)
		}

		got := matchArgsAgainstURLValues(tc.args, values)

		if got != tc.want {
			t.Errorf("matchArgsAgainstURLValues(%v, %s): Got %v, want %v", tc.args, values, got, tc.want)
		}
	}
}
