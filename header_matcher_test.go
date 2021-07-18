package hex

import (
	"fmt"
	"net/http/httptest"
)

func ExampleExpectation_WithHeader() {
	e := Expecter{}

	e.ExpectReq("GET", "/foo").WithHeader("Authorization", R("^Bearer .+$"))

	req := httptest.NewRequest("GET", "/foo", nil)
	req.Header.Set("Authorization", "Bearer foobar")

	e.LogReq(req)

	fmt.Println(e.Summary())
	// Output:
	// Expectations
	//	GET /foo with header matching Authorization="^Bearer .+$" - passed
}
