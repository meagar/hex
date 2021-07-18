package hex

import (
	"io"
	"net/http"
)

// RespondWithHandler registers an alternate handler to use when the expectation matches a request.
// Use AndCallThrough to additionally run the original handler, after the new handler is called
func (e *Expectation) RespondWithHandler(handler http.Handler) *Expectation {
	if e.handler != nil {
		panic("Multiple responses defined for one hex.Expectation")
	}
	e.handler = handler
	return e
}

// RespondWithFn adds a mock response using a function that can be passed to http.HandlerFunc
func (e *Expectation) RespondWithFn(fn func(http.ResponseWriter, *http.Request)) *Expectation {
	return e.RespondWithHandler(http.HandlerFunc(fn))
}

// RespondWith accepts a status code and string respond body
func (e *Expectation) RespondWith(status int, body string) *Expectation {
	return e.RespondWithFn(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(status)
		if _, err := io.WriteString(rw, body); err != nil {
			panic("Failed to write response in RespondWith")
		}
	})
}

// AndCallThrough instructs the expectation to run both a registered mock response handler, and then
// additionally run the original handler
func (e *Expectation) AndCallThrough() *Expectation {
	if e.handler == nil {
		panic("AndCallThrough called on expectation that has no mock response. Use WithResponse to setup a mock response before calling AndCallThrough")
	}
	e.callThrough = true
	return e
}
