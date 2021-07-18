package hex

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
)

func matchArgsAgainstURLValues(args []interface{}, values url.Values) bool {
	// With no arguments, we simply test whether the values are non-empty
	if len(args) == 0 {
		return len(values) > 0
	}

	if len(args) == 1 {
		return matchingParamsBySingle(args[0], values)
	}

	if len(args) == 2 {
		return matchingParamsByDouble(args[0], args[1], values)
	}

	return false
}

type matcher interface {
	matches(req *http.Request) bool
	String() string
}

func regexpOrStringMatcher(value interface{}) func(string) bool {
	if str, ok := value.(string); ok {
		return func(val string) bool {
			return val == str
		}
	} else if re, ok := value.(*regexp.Regexp); ok {
		return func(val string) bool {
			return re.MatchString(val)
		}
	} else {
		panic("Unknown value type")
	}
}

type valueMatcher interface {
	match(value interface{}) bool
}

// MatchConst defines several constants for matching anything
type MatchConst int

var (
	// Any matches any value
	Any MatchConst = 1

	// AnyString matches any string value
	AnyString MatchConst = 2
)

var _ valueMatcher = Any

func (m MatchConst) match(value interface{}) bool {
	switch m {
	case Any:
		return true
	case AnyString:
		_, ok := value.(string)
		return ok
	default:
		panic("Unknown MatchConst value")
	}
}

type stringMatcher struct {
	str string
}

func (s stringMatcher) match(value interface{}) bool {
	if str, ok := value.(string); ok {
		return str == s.str
	}

	return false
}

var _ valueMatcher = stringMatcher{}

type regexpMatcher struct {
	regex *regexp.Regexp
}

func (s regexpMatcher) match(value interface{}) bool {
	if str, ok := value.(string); ok {
		return s.regex.MatchString(str)
	}
	// TODO: Should ints, floats be matched?
	return false
}

var _ valueMatcher = regexpMatcher{}

func buildMatcher(value interface{}) valueMatcher {
	if str, ok := value.(string); ok {
		return stringMatcher{str}
	} else if regexp, ok := value.(*regexp.Regexp); ok {
		return regexpMatcher{regexp}
	} else if valueMatcher, ok := value.(valueMatcher); ok {
		return valueMatcher
	}

	log.Panicf("Unknown value matcher type %v", value)
	return nil
}

// matchingParamsByKey returns a new url.Values containing all matched key/value pairs
func matchingParamsBySingle(key interface{}, values url.Values) bool {
	matchMap := make(map[valueMatcher]valueMatcher)
	if str, ok := key.(string); ok {
		matchMap[stringMatcher{str}] = Any
	} else if regexp, ok := key.(*regexp.Regexp); ok {
		matchMap[regexpMatcher{regexp}] = Any
	} else if p, ok := key.(P); ok {
		for key, value := range p {
			matchMap[buildMatcher(key)] = buildMatcher(value)
		}
	} else if m, ok := key.(valueMatcher); ok {
		// TODO: Does this work?
		matchMap[m] = Any
	} else {
		log.Panicf("Unable to interpret argument %v as a type of matcher", key)
	}

	return performMatch(matchMap, values)
}

func matchingParamsByDouble(key interface{}, value interface{}, values url.Values) bool {
	matchMap := map[valueMatcher]valueMatcher{
		buildMatcher(key): buildMatcher(value),
	}
	return performMatch(matchMap, values)
}

func performMatch(matchMap map[valueMatcher]valueMatcher, values url.Values) bool {
	for matchKey, matchValue := range matchMap {
		matched := false
		for candidateKey, candidateValues := range values {
			for _, candidateValue := range candidateValues {
				if matchKey.match(candidateKey) && matchValue.match(candidateValue) {
					matched = true
					break
				}
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func matcherArgsToString(args []interface{}) string {
	if len(args) == 0 {
		return "<no args>"
	}

	if len(args) == 1 {
		return fmt.Sprintf("%s", args[0])
	}

	if len(args) == 2 {
		return fmt.Sprintf("%s=%q", args[0], args[1])
	}

	panic("Too many arguments")
}
