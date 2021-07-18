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

func TestQueryMatcher(t *testing.T) {
	e := Expecter{}
	e.ExpectReq("GET", "/foo").WithQuery("name", "bob")
	e.LogReq(httptest.NewRequest("GET", "/foo?name=bob", nil))
	if !e.Pass() {
		t.Errorf("WithQuery(\"name\", \"bob\") did not match ?name=bob")
	}
}
