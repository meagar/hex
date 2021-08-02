package hex

import (
	"fmt"
	"net/http"
	"net/url"
)

// WithHeader adds matching conditions against a request's headers
func (exp *Expectation) WithHeader(args ...interface{}) *Expectation {
	matcher, err := makeURLValuesMatcher(args)
	if err != nil {
		panic(fmt.Sprintf("WithHeader: %s", err.Error()))
	}
	exp.matchers = append(exp.matchers, &headerMatcher{
		args:             args,
		urlValuesMatcher: matcher,
	})
	return exp
}

type headerMatcher struct {
	args []interface{}
	urlValuesMatcher
}

var _ matcher = &headerMatcher{}

func (h *headerMatcher) matches(req *http.Request) bool {
	return h.urlValuesMatcher.matches(url.Values(req.Header))
}

func (h *headerMatcher) String() string {
	return fmt.Sprintf("header matching %v", matcherArgsToString(h.args))
}
