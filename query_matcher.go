package hex

import (
	"fmt"
	"net/http"
)

// WithQuery matches against the query string.
// It has several forms:
//
//   WithQuery() // passes if any query string is present
//   WithQuery("key") // passes ?key and ?key=<any value>
//   WithQuery("key", "value") // passes ?key=value or &key=value&key=value2
//   WithQuery(hex.R(`^key$`)) // find keys using regular expressions
//   WithQuery(hex.P{"key1": "value1", "key2": "value2"}) // match against multiple key/value pairs
//   WithQuery(hex.P{"key1": hex.R(`^value\d$`)}) // mix-and-match strings, regular expressions and key/value maps
func (exp *Expectation) WithQuery(args ...interface{}) *Expectation {
	matcher, err := makeURLValuesMatcher(args)
	if err != nil {
		panic(fmt.Sprintf("WithQuery: %s", err.Error()))
	}

	exp.matchers = append(exp.matchers, &queryMatcher{
		args:             args,
		urlValuesMatcher: matcher,
	})
	return exp
}

type queryMatcher struct {
	args  []interface{}
	exact bool
	urlValuesMatcher
}

func (q *queryMatcher) matches(req *http.Request) bool {
	query := req.URL.Query()
	return q.urlValuesMatcher.matches(query)

}

func (q *queryMatcher) String() string {
	return fmt.Sprintf("query string matching %s", matcherArgsToString(q.args))
}

var _ matcher = &queryMatcher{}
