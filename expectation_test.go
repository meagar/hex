package hex

import (
	"fmt"
	"net/http/httptest"
	"testing"
)

func ExampleExpectation_Never() {
	e := Expecter{}

	e.ExpectReq("GET", "/users").Never()
	e.LogReq(httptest.NewRequest("GET", "/users", nil))

	fmt.Println(e.Summary())
	// Output:
	// Expectations
	// 	GET /users - failed, expected 0 matches, got 1
}

func ExampleExpectation_Once() {
	e := Expecter{}

	e.ExpectReq("GET", "/status").Once()

	e.LogReq(httptest.NewRequest("GET", "/status", nil))
	e.LogReq(httptest.NewRequest("GET", "/status", nil))

	fmt.Println(e.Summary())
	// Output:
	// Expectations
	// 	GET /status - failed, expected 1 matches, got 2
}

func TestExpectation(t *testing.T) {
	testCases := []struct {
		expMethod  string
		expPath    string
		reqMethod  string
		reqPath    string
		shouldPass bool
	}{
		// Cases that should match
		{"GET", "/foobar", "GET", "/foobar", true},
		{"POST", "/foo/baz", "POST", "/foo/baz", true},
		{"PATCH", "/", "PATCH", "/", true},

		// Cases that should not match
		{"GET", "/foobar", "POST", "/foobar", false},
		{"GET", "/foobar", "GET", "/foobar/1", false},
		{"PATCH", "/", "PATCH", "/foobar", false},
	}

	for _, tc := range testCases {
		e := Expecter{}

		e.ExpectReq(tc.expMethod, tc.expPath).Do(func() {
			e.LogReq(httptest.NewRequest(tc.reqMethod, tc.reqPath, nil))
		})

		if tc.shouldPass {
			if !e.Pass() {
				t.Errorf("Expected request %s %s to match expectation %s %s, but it didn't",
					tc.reqMethod, tc.reqPath, tc.expMethod, tc.expPath)
			}
		} else {
			if e.Pass() {
				t.Errorf("Expected request %s %s to not match expectation %s %s, but it did",
					tc.reqMethod, tc.reqPath, tc.expMethod, tc.expPath)

			}
		}
	}
}

func TestNever(t *testing.T) {
	t.Run("Never() passes when a expectation does not match any requests", func(t *testing.T) {
		e := Expecter{}
		e.ExpectReq("GET", "/foo").Never().Do(func() {
			e.LogReq(httptest.NewRequest("GET", "/bar", nil))
		})

		if !e.Pass() {
			t.Errorf("Never() failed when it was expected to pass")
		}
	})

	t.Run("Never() fails when a expectation matches a request", func(t *testing.T) {
		e := Expecter{}
		e.ExpectReq("GET", "/foo").Never().Do(func() {
			e.LogReq(httptest.NewRequest("GET", "/foo", nil))
		})

		if e.Pass() {
			t.Errorf("Never() passed when it was expected to fail")
		}
	})

	t.Run("Never() causes a failure when a request with sub-conditions matches", func(t *testing.T) {
		e := Expecter{}
		e.ExpectReq("GET", "/foo").WithQuery("name", "bob").Never().Do(func() {
			e.LogReq(httptest.NewRequest("GET", "/foo?name=bob&foo=bar", nil))
		})

		if e.Pass() {
			t.Errorf("Never() passed when it was expected to fail")
		}
	})

	t.Run("Never() fails when an expectation matches in method/path but differs in sub-conditions", func(t *testing.T) {
		e := Expecter{}
		e.ExpectReq("GET", "/foo").WithQuery("name", "bob").Never().Do(func() {
			e.LogReq(httptest.NewRequest("GET", "/foo?name=sam", nil))
		})

		if !e.Pass() {
			t.Errorf("Never() failed when it was expected to pass")
		}
	})
}

func TestOnce(t *testing.T) {
	t.Run("Once() passes when a expectation matches one request", func(t *testing.T) {
		e := Expecter{}
		e.ExpectReq("GET", "/foo").Once().Do(func() {
			e.LogReq(httptest.NewRequest("GET", "/foo", nil))
		})

		if !e.Pass() {
			t.Errorf("Once() failed when it was expected to pass")
		}
	})

	t.Run("Once() fails when a expectation does not match a request", func(t *testing.T) {
		e := Expecter{}
		e.ExpectReq("GET", "/foo").Once().Do(func() {
			e.LogReq(httptest.NewRequest("GET", "/bar", nil))
		})

		if e.Pass() {
			t.Errorf("Once() passed when it was expected to fail")
		}
	})

	t.Run("Once() fails when a expectation matches more than one request", func(t *testing.T) {
		e := Expecter{}
		e.ExpectReq("GET", "/foo").Once().Do(func() {
			e.LogReq(httptest.NewRequest("GET", "/foo", nil))
			e.LogReq(httptest.NewRequest("GET", "/foo", nil))
		})

		if e.Pass() {
			t.Errorf("Once() passed when it was expected to fail")
		}
	})

	t.Run("Once() passes when a request with sub-conditions matches", func(t *testing.T) {
		e := Expecter{}
		e.ExpectReq("GET", "/foo").WithQuery("name", "bob").Once().Do(func() {
			e.LogReq(httptest.NewRequest("GET", "/foo?name=bob&foo=bar", nil))
		})

		if !e.Pass() {
			t.Errorf("Once() failed when it was expected to pass")
		}
	})

	t.Run("Once() fails when an expectation matches in method/path but differs in sub-conditions", func(t *testing.T) {
		e := Expecter{}
		e.ExpectReq("GET", "/foo").WithQuery("name", "bob").Once().Do(func() {
			e.LogReq(httptest.NewRequest("GET", "/foo?name=sam", nil))
		})

		if e.Pass() {
			t.Errorf("Once() passed when it was expected to fail")
		}
	})
}
