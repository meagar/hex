package hex

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"strings"
)

func ExampleExpectation_WithBody() {
	e := Expecter{}

	e.ExpectReq("POST", "/posts").WithBody("title", "My first blog post").Once()

	body := strings.NewReader(url.Values{
		"title": []string{"My first blog post"},
	}.Encode())
	req := httptest.NewRequest("POST", "/posts", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	e.LogReq(req)

	fmt.Println(e.Summary())
	// Output:
	// Expectations
	// 	POST /posts with body matching title="My first blog post" - passed
}
