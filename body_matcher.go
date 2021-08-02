package hex

import (
	"fmt"
	"net/http"
)

// WithBody adds matching conditions against a request's body.
// See WithQuery for usage instructions
func (e *Expectation) WithBody(args ...interface{}) *Expectation {
	matcher, err := makeURLValuesMatcher(args)
	if err != nil {
		panic(fmt.Sprintf("WithBody: %s", err.Error()))
	}

	e.matchers = append(e.matchers, &bodyMatcher{
		args:             args,
		urlValuesMatcher: matcher,
	})
	return e
}

type bodyMatcher struct {
	args             []interface{}
	urlValuesMatcher urlValuesMatcher
}

var _ matcher = &bodyMatcher{}

func (b *bodyMatcher) matches(req *http.Request) bool {
	if err := req.ParseForm(); err != nil {
		panic("An error occurred while parsing a form")
	}

	return b.urlValuesMatcher.matches(req.PostForm)
}

func (b *bodyMatcher) String() string {
	return fmt.Sprintf("body matching %v", matcherArgsToString(b.args))
}
