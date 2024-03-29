# Hex - Http EXpectations

Hex is a simple wrapper that extends `httptest.Server` with an expectation syntax, allowing you to create mock APIs using a simple and expressive DSL:


```go
func TestUserClient(t*testing.T) {
	// A mock of some remote user service
	server := hex.NewServer(t, nil)	 // nil, or optional http.Handler

	server.ExpectReq("GET", "/users").
		WithHeader("Authorization", "Bearer xxyyzz").
		WithQuery("search", "foo").
		RespondWith(200, `{"id": 123, "name": "test_user"}`)

	// Actual client implementation would go here
	http.Get(server.URL + "/foo")

	// Output:
	// example_test.go:12: One or more HTTP expectations failed
	// example_test.go:12: Expectations
	// example_test.go:12: 	GET /users with header matching Authorization="Bearer xyz"query string matching search="foo" - failed, no matching requests
	// example_test.go:12: Unmatched Requests
	// example_test.go:12: 	GET /foo
}
```

## Getting Started

Hex  provides a higher level `Server` which embeds `http.Server`. Create one with `hex.NewServer`, and start making expectations.
The arguments to `hex.NewServer` are a `*testing.T`, and an optional `http.Handler` which may be `nil`.
If an existing `http.Handler` is passed to `NewServer`, hex will pass requests to it *after* checking them against its tree of expectations.

```go
s := hex.NewServer(t, http.HandlerFunc(rw ResponseWriter, req *http.Request) {
	fmt.Fprintf(tw, "Ok")
})

s.ExpectReq("GET", "/users").WithQuery("page", "1")

http.Get(s.URL + "/users?page=1") // Match
```

If you have an existing mock, it can embed an `hex.Expecter`, which provides `ExpectReq` for setting up expectations, and `LogReq` for logging incoming requests so they can be matched against expectations. [`Server`](https://github.com/meagar/hex/blob/main/server.go) does exactly this, and serves an an example of how to write up the necessary plumbing.

## Matching Requests

Expectations are setup via `ExpectReq`, which accepts an HTTP method (one of `"GET"`, `"POST"`, `"PATCH"`, etc) and a path (not including query string):

```go
server.ExpectReq("GET", "/path/to/resource")
```

`ExpectReq` accepts one or two `interface{}` values, where each value is one of the following:

-  A string, ie `"GET"` or `"/path/to/resource`
- A regular expression created via `regexp.MustCompile` or the convenience method `hex.R`
- A built-in matcher like `hex.Any` or `hex.None`
- A function of type `hex.MatchFn` (`func(req*http.Request) bool`)
- A `map[interface{}]interface{}` which can recursively contain any of the above (typically only useful for matching against header/body/query string)

### Reporting Failure

hex will automatically report failures, and let you know which HTTP requests were made that didn't match any expectations:

```go
func TestExample(t *testing.T) {
	s := hex.NewServer(t, nil)

	s.ExpectReq("GET", "/foo")

	http.Get(s.URL + "/bar")
}
```

Output:

```plain
$ go test ./example
--- FAIL: TestExample (0.00s)
    server.go:29: One or more HTTP expectations failed
    print.go:205: Expectations
    print.go:205: 	GET /foo - failed, no matching requests
    print.go:205: Unmatched Requests
    print.go:205: 	GET /bar
FAIL
FAIL	github.com/meagar/hex/example	0.260s
FAIL
```

### Matching against strings, regular expressions, functions and more

Any key or value given to `ExpectReq`, `WithQuery`, `WithHeader` or `WithBody` can one of:

- A string, in which case case-sensitive exact matching is used:

	```go
	server.ExpectReq("GET", "/users") // matches GET /users?foo=bar
	```

* A regular expression (via `regexp.MustCompile` or `hex.R`):

	```go
	server.ExpectReq(hex.R("^(POST|PATCH)$", hex.R("^users/\d+$")
	```

* One of several predefined constants like `hex.Any` or `hex.None`

	```go
	server.ExpectReq("GET", hex.Any)                             // matches any GET request
	server.ExpectReq(hex.Any, hex.Any)                           // matches *any* request
	server.ExpectReq(hex.Any, hex.Any).WithQuery(hex.Any, "123") // Matches any request with any query string parameter having the value "123"
	```

* A map of `interface{}`/`interface{}` pairs, where each `interface{}` value is itself a string/regex/map/

	```go
	server.ExpectReq("GET", "/search").WithQuery(hex.P{
		"q": "test",
		"page": hex.R(`^\d+$`),
	})
	```

### Matching against the query string, header and body

You can make expectations about the query string, headers or form body with `WithQuery`, `WithHeader` and `WithBody` respectively:

```go
func TestClientLibrary(t*testing.T) {
	t.Run("It includes the authorization header", func(t*testing.T) {
		server := hex.NewServer(t, nil)
		server.ExpectReq("GET", "/users").WithHeader("Authorization", hex.R("^Bearer .+$"))
		// ...
		client.GetUsers()
	})

	t.Run("It includes the user Id in the query string", func(t*testing.T) {
		server := hex.NewServer(t, nil)
		server.ExpectReq("GET", "/users").WithQuery("id", "123")
		// ...
		client.GetUser("123")
	})
}
```

When only one argument is given to any `With*` method, matching is done against the key, with any value being accepted:toc:

```go
server.ExpectReq("GET", "/users").WithQuery("id")
// ...
http.Get(server.URL + "/users")              // fail
http.Get(server.URL + "/users?id")           // pass
http.Get(server.URL + "/users?id=1")         // pass
http.Get(server.URL + "/users?id=1&foo=bar") // pass
```

When no arguments are given, `WithQuery`, `WithHeader` and `WithBody` match any request with a non-empty query/header/body respectively.

```go
server.ExpectReq("GET", "/users").WithQuery()
// ...
http.Get(server.URL + "/users")         // fail
http.Get(server.URL + "/users?foo")     // pass
http.Get(server.URL + "/users?foo=bar") // pass
http.Get(server.URL + "/users?foo=bar") // pass
```

## Mocking Responses

By default, hex will pass requests to the `http.Handler` object you provide through `NewServer` (if any).
You can override the response with `RespondWith(status int, body string)`, `ResponseWidthFn(func(http.ResponseWriter, *http.Request))` or `RespondWithHandler(http.Handler)`:

```go
server := hex.NewServer(t, nil)
server.ExpectReq("GET", "/users").RespondWith("200", "OK")
```

By default, the `http.Handler` you provide to `NewServer` will not be invoked if a requests matches an expectation for which a mock response has been defined.
However, you can allow the request to "fall through" and reach your own handler with `AndCallThrough`.
Note that, if your handler writes a response, it will be concatenated to the mock response already produced, and any HTTP status you attempt to write will be silently discarded  if a mock response has already set one.:

```go
server := hex.NewServer(t, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf("BBB")
}))

// Requests matching this expectation will receive a response of "AAABBB"
server.ExpectReq("GET", "/foo").RespondWith(200, "AAA").AndCallThrough()
```

## Scoping with `Do`

By default, a request issued at any point in a test after an `ExpectReq` expectation is made will match that expectation.

To limit the scope in which an expectation can be matched, use `Do`:

```go
server := hex.NewServer(t, nil)
server.ExpectReq("GET", "/users").Do(func() {
	// This will match:
	http.Get(server.URL + "/users")
})
// This will fail, the previous expectation's scope has closed
http.Get(server.URL + "/users")
```

## `Once`, `Never`

If a request should only happen once (or not at all) in a given block of code, you can express this expectation with `Once` or `Never`:

```go
func TestCaching(t*testing.T) {
	t.Run("The client caches the server's response", func(t*testing.t) {
		server := hex.NewServer(t, nil)
		server.ExpectReq("GET", "/countries").Once()
		// ...
		client.GetCountries()
		client.GetCountries()
		// Output:
		// Expectations
		// 	GET /countries - failed, expected 1 matches, got 2
	})

	t.Run("The client should not make a request if the arguments are invalid", func(t*testing.T) {
		server := hex.NewServer(t, nil)
		server.ExpectReq("GET", "/users").Never()
		// ...
		// Assume the client is not supposed to make requests unless the ID is an integer
		_, err := client.GetUser("foo")
		// assert that err is not nil
	})
})
```

## Helpers `R` and `P`

`hex.R` is a wrapper around `regexp.MustCompile`, and `hex.P` ("params") is an alias for `map[string]interface{}`.

These helpers allow for more succinct definition of matchers:

```go
server := hex.NewServer(t, nil)
server.ExpectReq("GET", hex.R(`/users/\d+`)) // Matches /users/123
// ... 
server.ExpectReq("POST", "/users").WithBody(hex.P{
	"name": hex.R(`^[a-z]+$`),
	"age": hex.R(`^\d+$`),
})
```

## TODO

- [ ] Better support for matching JSON requests
- [ ] Higher level helpers
	- [ ] `WithBearer`
	- [ ] `WithJsonResponse`
	- [ ] `WithType("json"|"html")`
- [ ] `hex.Verbose()` and `ExpectReq(...).Verbose()` for debugging
