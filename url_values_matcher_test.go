package hex

import (
	"fmt"
	"log"
	"net/url"
	"testing"
)

func TestURLValuesMatcher_Match(t *testing.T) {

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

		// matchers using functions
		// {Args{func(interface{}) bool { return true }}, "foo=bar", true},
		// {Args{func(interface{}) bool { return false }}, "foo=bar", false},

		// matching against param maps with string/regex keys and values
		{Args{P{"key1": "value1", "key2": "value2"}}, "key1=value1&key2=value2", true},
		{Args{P{"key1": R(`^value\d+$`), "key2": R(`^value\d+$`)}}, "key1=value1&key2=value1", true},
		{Args{P{"key1": R(`^value\d+$`), "key2": R(`^value\d+$`)}}, "key1=value&key2=value", false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("matchArgsAgainstURLValues(%v, %s)", tc.args, tc.encodedParams), func(t *testing.T) {
			values, err := url.ParseQuery(tc.encodedParams)
			if err != nil {
				panic("Invalid query string:" + tc.encodedParams)
			}

			u, err := makeURLValuesMatcher(tc.args)
			if err != nil {
				log.Panic("Unexpected error creating urlValuesMatcher:", err)
			}
			got := u.matches(values)

			if got != tc.want {
				t.Errorf("matchArgsAgainstURLValues(%v, %s): Got %v, want %v", tc.args, values, got, tc.want)
				t.FailNow()
			}
		})
	}

}
