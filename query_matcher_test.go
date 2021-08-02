package hex

import (
	"fmt"
	"net/http/httptest"
	"testing"
)

func ExampleExpectation_WithQuery() {
	e := Expecter{}

	e.ExpectReq("GET", "/search").WithQuery("q", "cats")

	e.LogReq(httptest.NewRequest("GET", "/search?q=cats", nil))

	fmt.Println(e.Summary())
	// Output:
	// Expectations
	//	GET /search with query string matching q="cats" - passed
}

func ExampleExpectation_WithQuery_invalidArgument() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Panic:", err)
		}
	}()

	e := Expecter{}
	// Unrecognized arguments to WithQuery produce a panic.
	e.ExpectReq("GET", "/search").WithQuery(123)

	// Output:
	// Panic: WithQuery: Cannot use value 123 when matching against url.Values
}

func TestQueryMatcher(t *testing.T) {
	e := Expecter{}
	e.ExpectReq("GET", "/foo").WithQuery("name", "bob")
	e.LogReq(httptest.NewRequest("GET", "/foo?name=bob", nil))
	if !e.Pass() {
		t.Errorf("WithQuery(\"name\", \"bob\") did not match ?name=bob")
	}
}
