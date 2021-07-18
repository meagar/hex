package hex

import "regexp"

// R is a convenience wrapper for regexp.MustCompile
func R(pattern string) *regexp.Regexp {
	return regexp.MustCompile(pattern)
}

// P is a convenience alias for a map of interface{} to interface{}
//
// It's used to add header/body/query string conditions to an expectation:
//
//   server.Expect("GET", "/foo").WithQuery(hex.P{"name": "bob", "age": hex.R("^\d+$")})
type P map[interface{}]interface{}
