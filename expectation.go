package hex

import (
	"fmt"
	"net/http"
	"strings"
)

// Expectation captures details about an ExpectReq call and subsequent conditions
// chained to it.
type Expectation struct {
	// Our parent expecter object
	expecter *Expecter

	parent   *Expectation
	children []*Expectation

	method string
	path   string

	quantifier *quantifier

	matches  []*http.Request
	matchers []matcher

	debug bool

	handler     http.Handler
	callThrough bool
}

type quantifier struct {
	desc            string
	min, max, count uint
}

func (e *Expectation) String() string {
	buf := &strings.Builder{}
	fmt.Fprintf(buf, "%s %s", e.method, e.path)
	if len(e.matchers) > 0 {
		fmt.Fprintf(buf, " with ")
		for _, m := range e.matchers {
			fmt.Fprintf(buf, m.String())
		}
	}

	if e.pass() {
		fmt.Fprintf(buf, " - passed")
	} else {
		fmt.Fprintf(buf, " - failed, %s", e.failureReason())

	}

	return buf.String()
}

func (e *Expectation) failureReason() string {
	if e.pass() {
		panic("failureReason called for non-failing expectation")
	}

	if len(e.matches) == 0 {
		return "no matching requests"
	}

	if e.quantifier != nil {
		if e.quantifier.min == e.quantifier.max {
			return fmt.Sprintf("expected %d matches, got %d", e.quantifier.min, e.quantifier.count)
		}
		return fmt.Sprintf("expected %d..%d matches, got %d", e.quantifier.min, e.quantifier.max, e.quantifier.count)
	}
	return ""
}

// Do opens a scope. Expectations in the current scope may be matched by requests in the current or nested scopes, but
// requests in higher scopes cannot fulfill expections in lower scopes.
//
// For example:
//
//     expector.ExpectReq("POST", "/foo")
//     expector.ExpectReq("GET", "/bar").Do(func() {
//       // matches POST expectation in parent scope
//       expector.LogReq(httptest.NewRequest("GET" "/foo", nil))
//     })
//
//     // Does NOT match GET expectation in previous scope
//     expector.LogReq(httptest.NewRequest("GET" "/foo", nil)) // does not match
//
// The current expectation becomes the first expectation within the new scope
func (e *Expectation) Do(fn func()) {
	e.expecter.do(fn)
}

func (e *Expectation) pass() bool {
	if e.quantifier != nil {
		return e.quantifier.count >= e.quantifier.min && e.quantifier.count <= e.quantifier.max
	}

	return len(e.matches) > 0
}

// Matches returns true if the expectation is fulfilled by the given http.Request
// TODO: Support header, body, etc.
func (e *Expectation) matchAgainst(req *http.Request) bool {
	// Baseline check against method and path
	if req.Method != e.method || req.URL.Path != e.path {
		return false
	}

	for _, c := range e.matchers {
		if !c.matches(req) {
			return false
		}
	}

	e.matches = append(e.matches, req)
	if e.quantifier != nil {
		e.quantifier.count++
	}

	if e.debug {
		fmt.Println("Debug")
		fmt.Println(req)
		fmt.Println(req.Header)
		fmt.Println(req.URL.Query())
		fmt.Println(req.Form)
		panic("debug")
	}

	return true
}

// Quantification

func (e *Expectation) quantify(desc string, min, max uint) {
	if e.quantifier != nil {
		panic("A quantifier was added multiple times to the same expectations")
	}

	e.quantifier = &quantifier{
		desc: desc,
		min:  min,
		max:  max,
	}
}

// Never asserts that the expectation is matched zero times
func (e *Expectation) Never() *Expectation {
	e.quantify("never", 0, 0)
	return e
}

// Once adds a quantity condition that requires exactly one request to be matched
func (e *Expectation) Once() *Expectation {
	e.quantify("once", 1, 1)
	return e
}

// Debug marks an expectation for debugging. When it matches a request, it will dump details about
// the request to STDOUT
func (e *Expectation) Debug() *Expectation {
	e.debug = true
	return e
}

// func (exp *Expectation) AtMost(n uint) *Expectation {
// 	exp.quantify("at most", 0, n)
// 	return exp
// }

// func (exp *Expectation) NTimes(n uint) *Expectation {
// 	exp.quantify("exactly", n, n)
// 	return exp
// }

// func (exp *Expectation) AtLeast(n uint) *Expectation {
// 	exp.quantify("at least", n, math.MaxUint32)
// 	return exp
// }
