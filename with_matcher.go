package hex

import "net/http"

type withMatcher struct {
	fn func(*http.Request) bool
}

var _ matcher = &withMatcher{}

func (w *withMatcher) matches(req *http.Request) bool {
	return w.fn(req)
}

func (w *withMatcher) String() string {
	return "<custom With matcher>"
}

// With adds a generic condition callback that must return true if the request matched, and false otherwise
func (e *Expectation) With(fn func(req *http.Request) bool) {
	e.matchers = append(e.matchers, &withMatcher{fn: fn})
}
