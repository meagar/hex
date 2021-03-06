package hex

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"
)

// Expecter is the top-level object onto which expectations are made
type Expecter struct {
	root    *Expectation
	current *Expectation

	matched   []*http.Request
	unmatched []*http.Request
}

// Pass returns true if all expectations have passed
func (e *Expecter) Pass() bool {
	return len(e.FailedExpectations()) == 0
}

// Fail returns true if any expectation has failed
func (e *Expecter) Fail() bool {
	return !e.Pass()
}

// UnmatchedRequests returns a list of all http.Request objects that didn't match any expectation
func (e *Expecter) UnmatchedRequests() []*http.Request {
	return e.unmatched
}

// PassedExpectations returns all passing expectations
func (e *Expecter) PassedExpectations() (passed []*Expectation) {
	if e.root == nil {
		return
	}

	var findPasses func(exp *Expectation)
	findPasses = func(exp *Expectation) {
		if exp != e.root && exp.pass() {
			passed = append(passed, exp)
		}
		for _, child := range exp.children {
			findPasses(child)
		}
	}
	findPasses(e.root)
	return
}

// FailedExpectations returns a list of currently failing
func (e *Expecter) FailedExpectations() (failed []*Expectation) {
	if e.root == nil {
		return
	}

	var findFailures func(exp *Expectation)
	findFailures = func(exp *Expectation) {
		if exp != e.root && !exp.pass() {
			failed = append(failed, exp)
		}
		for _, child := range exp.children {
			findFailures(child)
		}
	}
	findFailures(e.root)
	return
}

// ExpectReq adds an Expectation to the stack
func (e *Expecter) ExpectReq(method, path interface{}) (exp *Expectation) {
	if e.root == nil {
		// Lazily initialize the Expecter, so the zero-value is usable
		e.root = &Expectation{
			method: &noStringMatcher{},
			path:   &noStringMatcher{},
		}

		e.current = e.root
	}

	methodMatcher, err := makeStringMatcher(method)
	if err != nil {
		log.Panicf("Invalid HTTP method matcher %v in ExpectReq: %s", method, err.Error())
	}

	pathMatcher, err := makeStringMatcher(path)
	if err != nil {
		log.Panicf("Invalid HTTP path matcher %v in ExpectReq: %s", path, err.Error())
	}

	exp = &Expectation{
		method:   methodMatcher,
		path:     pathMatcher,
		expecter: e,
		parent:   e.current,
	}

	e.current.children = append(e.current.children, exp)
	e.current = exp

	return
}

// LogReq matches an incoming request against he current tree of Expectations, and returns the matched Expectation if any
func (e *Expecter) LogReq(req *http.Request) *Expectation {
	// Ascend up the stack, looking for expectations that match the given request
	var matched *Expectation
	for exp := e.current; exp != e.root; exp = exp.parent {
		if exp.matchAgainst(req) {
			matched = exp
		}
	}

	// When we reach the top level, we want to capture unmatched HTTP requests, so we can
	// display them in a report.
	if matched != nil {
		e.matched = append(e.matched, req)
	} else {
		e.unmatched = append(e.unmatched, req)
	}

	return matched
}

// do introduces a nested scope.
// When `do` is finished, the stack unwinds back to the expectation that introduced
// the scope, removing all nested expectation scopes.
func (e *Expecter) do(fn func()) {
	if e.current == nil {
		panic("Somehow `do` was invoked with an empty expectation stack")
	}

	current := e.current

	fn()

	e.current = current.parent
}

// TestingT covers the minimal interface we consume from a testing.T
type TestingT interface {
	Helper()
	Cleanup(func())
	Logf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

var _ TestingT = &testing.T{}

type captureT struct {
	buf strings.Builder
}

var _ TestingT = &captureT{}

func (t *captureT) Logf(format string, args ...interface{}) {
	fmt.Fprintf(&t.buf, format, args...)
}

func (t *captureT) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(&t.buf, format, args...)
}

func (t *captureT) Helper()        {}
func (t *captureT) Cleanup(func()) {}

// Summary returns a summary of all passed/failed expectations and any requests that didn't match
func (e *Expecter) Summary() string {
	t := captureT{}
	e.writeSummary(&t)
	return t.buf.String()
}

func (e *Expecter) writeSummary(t TestingT) {
	t.Helper()
	t.Logf("Expectations\n")
	for _, exp := range e.PassedExpectations() {
		t.Logf("\t%s\n", exp.String())
	}
	for _, exp := range e.FailedExpectations() {
		t.Logf("\t%s\n", exp.String())
	}

	if len(e.UnmatchedRequests()) > 0 {
		t.Logf("Unmatched Requests\n")
		for _, req := range e.UnmatchedRequests() {
			t.Logf("\t%s %s\n", req.Method, req.URL.Path)
		}
	}
}

// HexReport logs a summary of passes/fails to the given testing object, and calls t.Errorf with an error message if
// any expectations failed
func (e *Expecter) HexReport(t TestingT) {
	t.Helper()
	if e.Fail() {
		t.Errorf("One or more HTTP expectations failed\n")
	}

	e.writeSummary(t)
}
