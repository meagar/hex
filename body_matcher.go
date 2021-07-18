package hex

import (
	"fmt"
	"net/http"
)

// WithBody adds matching conditions against a request's body
func (e *Expectation) WithBody(args ...interface{}) *Expectation {
	e.matchers = append(e.matchers, &bodyMatcher{args: args})
	return e
}

type bodyMatcher struct {
	args []interface{}
}

var _ matcher = &bodyMatcher{}

func (b *bodyMatcher) matches(req *http.Request) bool {
	if err := req.ParseForm(); err != nil {
		panic("An error occurred while parsing a form")
	}

	return matchArgsAgainstURLValues(b.args, req.PostForm)
}

func (b *bodyMatcher) String() string {
	return fmt.Sprintf("body matching %v", matcherArgsToString(b.args))
}
