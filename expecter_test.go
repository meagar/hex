package hex

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func ExampleExpecter_Summary() {
	e := Expecter{}

	e.ExpectReq("GET", "/status")
	e.ExpectReq("POST", "/users")

	// Matches one of above expectations, leaving the other unmatched (failing)
	e.LogReq(httptest.NewRequest("GET", "/status", nil))

	// Extraneous request matches no expectations
	e.LogReq(httptest.NewRequest("PATCH", "/items", nil))

	fmt.Println(e.Summary())
	// Output:
	// Expectations
	// 	GET /status - passed
	// 	POST /users - failed, no matching requests
	// Unmatched Requests
	// 	PATCH /items
}
func TestExpecter(t *testing.T) {

	t.Run("With no expectations and no nesting, the Expecter passes", func(t *testing.T) {
		e := Expecter{}
		assertPassedFailed(t, 0, 0, e)
	})

	t.Run("requests logged before expectations do not match", func(t *testing.T) {
		e := Expecter{}
		e.LogReq(mockGet("/foo"))
		e.ExpectReq("GET", "/foo")
		assertPassedFailed(t, 0, 1, e)
	})

	t.Run("Expectations made at the top-level scope are tested by Done", func(t *testing.T) {
		t.Run("With one passing expectation", func(t *testing.T) {
			e := Expecter{}
			e.ExpectReq("GET", "/foobar")
			e.LogReq(mockGet("/foobar"))
			assertPassedFailed(t, 1, 0, e)
		})

		t.Run("With many passing expectations", func(t *testing.T) {
			e := Expecter{}
			e.ExpectReq("GET", "/foobar")
			e.ExpectReq("POST", "/foobar2")
			e.LogReq(mockGet("/foobar"))
			e.LogReq(mockPost("/foobar2", nil))
			assertPassedFailed(t, 2, 0, e)
		})

		t.Run("With one failing expectation", func(t *testing.T) {
			e := Expecter{}
			e.ExpectReq("GET", "/foobar")
			assertPassedFailed(t, 0, 1, e)
		})

		t.Run("With one failing and one passing expectation", func(t *testing.T) {
			e := Expecter{}
			e.ExpectReq("GET", "/foobar")
			e.ExpectReq("POST", "/foobar2")
			e.LogReq(mockPost("/foobar2", nil))
			assertPassedFailed(t, 1, 1, e)
		})
	})

	t.Run("When expectations introduce a scope, only requests in that scope match", func(t *testing.T) {
		t.Run("request in scope", func(t *testing.T) {
			e := Expecter{}

			e.ExpectReq("GET", "/foobar").Do(func() {
				e.LogReq(httptest.NewRequest("GET", "/foobar", nil))
			})

			assertPassed(t, 1, e)
		})

		t.Run("request out of scope", func(t *testing.T) {
			e := Expecter{}

			e.LogReq(httptest.NewRequest("GET", "/foobar", nil))
			e.ExpectReq("GET", "/foobar").Do(func() {
			})

			assertFailed(t, 1, e)
		})
	})

	t.Run("Requests that don't match an expectation are returned as unused", func(t *testing.T) {
		e := Expecter{}
		e.LogReq(mockGet("/abc"))
		e.ExpectReq("GET", "/foobar").Do(func() {
			e.LogReq(mockGet("/xyz"))
		})
		assertFailed(t, 1, e)
		assertUnused(t, e, mockGet("/xyz"), mockGet("/abc"))

	})

	t.Run("When simple nested expectations are met, the result is a pass", func(t *testing.T) {
		e := Expecter{}

		e.ExpectReq("GET", "/foobar").Do(func() {
			e.ExpectReq("POST", "/foobar2").Do(func() {
				e.LogReq(httptest.NewRequest("GET", "/foobar", nil))
				e.LogReq(httptest.NewRequest("POST", "/foobar2", nil))
			})
		})

		assertPassed(t, 2, e)
	})

	t.Run("When complex nested expectations are met, the result is a pass", func(t *testing.T) {
		e := Expecter{}

		e.ExpectReq("GET", "/foobar").Do(func() {
			e.ExpectReq("POST", "/foobar2").Do(func() {
				e.LogReq(httptest.NewRequest("GET", "/foobar", nil))
				e.LogReq(httptest.NewRequest("POST", "/foobar2", nil))
			})

			e.ExpectReq("PATCH", "/foobar3").Do(func() {
				e.ExpectReq("GET", "/foobar4").Do(func() {
					e.LogReq(httptest.NewRequest("GET", "/foobar4", nil))
					e.LogReq(httptest.NewRequest("PATCH", "/foobar3", nil))
				})
			})
		})

		assertPassed(t, 4, e)
	})

	t.Run("When an inner expectation would be matched by a request logged in an outer expectation, there is no match", func(t *testing.T) {
		e := Expecter{}

		e.ExpectReq("GET", "/foobar").Do(func() {
			e.LogReq(httptest.NewRequest("GET", "/foobar", nil))
			e.LogReq(httptest.NewRequest("POST", "/foobar2", nil))

			e.ExpectReq("POST", "/foobar2").Do(func() {
				// The request logged in the containing scope should not match this expectation
			})
		})

		assertPassedFailed(t, 1, 1, e)
	})

	t.Run("When one requests matches multiple expectations, a expectation are met", func(t *testing.T) {
		e := Expecter{}
		e.ExpectReq("GET", "/foobar").Do(func() {
			e.ExpectReq("GET", "/foobar").Do(func() {
				e.LogReq(mockGet("/foobar"))
			})
		})

		assertPassedFailed(t, 2, 0, e)
	})

	t.Run("When simply nested expectations are not met, the result is a fail", func(t *testing.T) {
		e := Expecter{}

		e.ExpectReq("GET", "/foobar").Do(func() {
			e.ExpectReq("POST", "/foobar2").Do(func() {
				e.LogReq(httptest.NewRequest("POST", "/foobar2", nil))
			})
		})

		assertPassedFailed(t, 1, 1, e)
	})

	t.Run("When complex nested expectations are not met, the result is a fail", func(t *testing.T) {
		e := Expecter{}

		e.ExpectReq("GET", "/foobar").Do(func() {
			e.ExpectReq("POST", "/foobar2").Do(func() {
			})
			e.ExpectReq("PATCH", "/foobar3").Do(func() {
				e.ExpectReq("GET", "/foobar4")
			})
		})

		assertFailed(t, 4, e)
	})

	t.Run("When complex nested expectations are partially met, the result is a fail", func(t *testing.T) {
		e := Expecter{}

		e.ExpectReq("GET", "/foobar").Do(func() {
			e.ExpectReq("POST", "/foobar2").Do(func() {
			})
			// This should *not* match the above expect, as it's not nested within it
			e.LogReq(mockPost("/foobar2", nil))
			e.ExpectReq("PATCH", "/foobar3").Do(func() {
				e.ExpectReq("GET", "/foobar4")
				// this should match, as the above expect does not introduce a scope
				e.LogReq(mockGet("/foobar4"))
			})
		})

		assertPassedFailed(t, 1, 3, e)
		assertUnused(t, e, mockPost("/foobar2", nil))
	})
}

func TestReport(t *testing.T) {
	mockT := TesterMock{}
	e := Expecter{}
	e.ExpectReq("GET", "/status")
	e.HexReport(&mockT)

	capturedOutput := mockT.b.String()
	expectedOutput := "One or more HTTP expectations failed\nExpectations\n\tGET /status - failed, no matching requests\n"
	if capturedOutput != expectedOutput {
		t.Errorf("Report wrote \n%s\n, expected \n%s\n", capturedOutput, expectedOutput)
	}
}

type TesterMock struct {
	b strings.Builder
}

func (t *TesterMock) Logf(format string, args ...interface{}) {
	fmt.Fprintf(&t.b, format, args...)
}

func (t *TesterMock) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(&t.b, format, args...)
}

func (t *TesterMock) Cleanup(func()) {

}

func (t *TesterMock) Helper() {
	// no-op
}

func mockGet(path string) *http.Request {
	return httptest.NewRequest("GET", path, nil)
}
func mockPost(path string, body io.Reader) *http.Request {
	return httptest.NewRequest("POST", path, body)
}

func assertUnused(t *testing.T, e Expecter, requests ...*http.Request) {
	fail := len(e.UnmatchedRequests()) != len(requests)

	if !fail {
		got := make([]*http.Request, len(e.UnmatchedRequests()))
		copy(got, e.UnmatchedRequests())
		for _, gotReq := range got {
			matched := false
			for wantIdx, wantReq := range requests {
				t.Log(requests)
				if gotReq.Method == wantReq.Method && gotReq.URL.Path == wantReq.URL.Path {
					matched = true
					requests = append(requests[:wantIdx], requests[wantIdx+1:]...)
					break
				}
			}
			if !matched {
				fail = true
				break
			}
		}
	}

	if fail {
		t.Error("Expected the following unused requests")
		for _, req := range requests {
			t.Errorf("%s %s\n", req.Method, req.URL.Path)
		}
		t.Error("Got the following unused requests")
		for _, req := range e.UnmatchedRequests() {
			t.Errorf("%s %s\n", req.Method, req.URL.Path)
		}
	}
}

func assertPassedFailed(t *testing.T, wantNumPassed, wantNumFailed int, e Expecter) {
	t.Helper()

	t.Log(e.Summary())

	for _, req := range e.UnmatchedRequests() {
		t.Logf("  %s %s", req.Method, req.URL.Path)
	}

	if wantNumFailed == 0 {
		if e.Pass() == false {
			t.Error("Pass() when all expectations were met: Got false, want true")
		}

		if e.Fail() == true {
			t.Error("Fail() when all expectations were met: Got true, want false")
		}
	} else {
		if e.Pass() == true {
			t.Error("Pass() when some expectations were not met: Got true, want false")
		}

		if e.Fail() == false {
			t.Error("Fail() when some expectations were not met: Got false, want true")
		}
	}

	gotNumPassed := len(e.PassedExpectations())
	if gotNumPassed != wantNumPassed {
		t.Errorf("len(Passed): Got %d, want %d", gotNumPassed, wantNumPassed)
	}

	gotNumFailed := len(e.FailedExpectations())
	if gotNumFailed != wantNumFailed {
		t.Errorf("len(Failed):  Got %d, want %d", gotNumFailed, wantNumFailed)
	}
}

func assertPassed(t *testing.T, numPassed int, e Expecter) {
	t.Helper()
	assertPassedFailed(t, numPassed, 0, e)
}

func assertFailed(t *testing.T, numFailed int, e Expecter) {
	t.Helper()
	assertPassedFailed(t, 0, numFailed, e)
}
