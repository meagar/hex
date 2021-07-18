package hex_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/meagar/hex"
)

func ExampleExpectation_With() {
	e := hex.Expecter{}

	// With allows custom matching through a callback function
	e.ExpectReq("POST", "/users").With(func(req *http.Request) bool {
		if err := req.ParseForm(); err != nil {
			panic(err)
		}
		return req.Form.Get("user_id") == "123"
	})

	body := strings.NewReader(url.Values{
		"user_id": []string{"123"},
	}.Encode())
	req := httptest.NewRequest("POST", "/users", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	e.LogReq(req)

	fmt.Println(e.Summary())
	// Output:
	// Expectations
	// 	POST /users with <custom With matcher> - passed
}
