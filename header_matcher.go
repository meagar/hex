package hex

import (
	"fmt"
	"net/http"
	"net/url"
)

// WithHeader adds matching conditions against a request's headers
func (exp *Expectation) WithHeader(args ...interface{}) *Expectation {
	exp.matchers = append(exp.matchers, &headerMatcher{args: args})
	return exp
}

type headerMatcher struct {
	args []interface{}
}

var _ matcher = &headerMatcher{}

func (h *headerMatcher) matches(req *http.Request) bool {
	return matchArgsAgainstURLValues(h.args, url.Values(req.Header))
}

func (h *headerMatcher) String() string {
	return fmt.Sprintf("header matching %v", matcherArgsToString(h.args))
}
