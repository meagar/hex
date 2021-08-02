package hex

import (
	"net/http"
	"net/http/httptest"
)

// Server wraps around (embeds) an httptest.Server, and also embeds an Expecter for making expectations
// The simplest way of using hex is to use NewServer or NewTLSServer.
type Server struct {
	*httptest.Server
	handler http.Handler
	t       TestingT
	Expecter
}

// NewServer returns a new hex.Server object, wrapping an httptest.Server.
// Its first argument should be a testing.T, used to report failures.
// Its second argument is an http.Handler that may be nil.
func NewServer(t TestingT, handler http.Handler) *Server {
	t.Helper()
	s := Server{
		t:       t,
		handler: handler,
	}
	s.Server = httptest.NewServer(&s)
	s.URL = s.Server.URL
	t.Cleanup(func() {
		t.Helper()
		s.HexReport(t)
	})

	return &s
}

// NewTLSServer returns a new hex.Server object, wrapping na httptest.Server created via NewTLSServer
// Its first argument should be a testing.T, used to report failures.
// Its second argument is an http.Handler that may be nil.
func NewTLSServer(t TestingT, handler http.Handler) *Server {
	t.Helper()
	s := Server{
		t:       t,
		handler: handler,
	}
	s.Server = httptest.NewTLSServer(&s)
	s.URL = s.Server.URL
	t.Cleanup(func() {
		t.Helper()
		s.HexReport(t)
	})

	x := new(string)
	*x = "foobar"

	return &s
}

// ServeHTTP logs requests that come through the server so they can be matched against expectations, and
// evalutes any mock responses defined for matched expectations.
func (s *Server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	exp := s.LogReq(req)

	if exp != nil && exp.handler != nil {
		exp.handler.ServeHTTP(rw, req)
		if exp.callThrough == false {
			return
		}
	}

	if s.handler != nil {
		s.handler.ServeHTTP(rw, req)
	}
}
