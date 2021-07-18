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
func (exp *Expectation) WithQuery(args ...interface{}) *Expectation {
	exp.matchers = append(exp.matchers, &queryMatcher{args: args})
	return exp
}

type queryMatcher struct {
	args  []interface{}
	exact bool
}

func (q *queryMatcher) matches(req *http.Request) bool {
	query := req.URL.Query()
	return matchArgsAgainstURLValues(q.args, query)

}

func (q *queryMatcher) String() string {
	return fmt.Sprintf("query string matching %s", matcherArgsToString(q.args))
}

var _ matcher = &queryMatcher{}
