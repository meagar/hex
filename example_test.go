package hex_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/meagar/hex"
)

// MockUserService is a mock we dependency-inject into our Client library
// It embeds a hex.Server so we can make HTTP requests of it, and use ExpectReq to set up
// HTTP expectations.
type MockUserService struct {
	*hex.Server
}

func NewMockUserService(t *testing.T) *MockUserService {
	s := MockUserService{}
	s.Server = hex.NewServer(t, &s)
	return &s
}

func (m *MockUserService) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	path := req.Method + " " + req.URL.Path

	switch path {
	case "GET /users":
		// TODO: Generate the response a client would expect
	case "POST /users":
		// TODO: Generate the response a client would expect
	default:
		rw.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(rw, "Not found")
	}
}

// SearchClient is our real HTTP client for the given service
type UsersClient struct {
	Host   string
	Client *http.Client
}

type User struct {
	Name  string
	Email string
}

// Search hits the "search" endpoint for the given host, with an id query string parameter
func (c *UsersClient) Find(userID int) (User, error) {
	_, err := c.Client.Get(fmt.Sprintf("%s/users/%d", c.Host, userID))
	// TODO: Decode mock service response
	return User{}, err
}

func (c *UsersClient) Create(u User) error {
	data := url.Values{}
	data.Set("name", u.Name)
	data.Set("email", u.Email)
	_, err := c.Client.PostForm(c.Host+"/users", data)
	return err
}

func Example() {
	t := testing.T{}
	service := NewMockUserService(&t)

	// Client is our real client implementation
	client := UsersClient{
		Client: service.Client(),
		Host:   service.URL,
	}

	// Make expectations about the client library
	service.ExpectReq("GET", "/users/123").Once().Do(func() {
		client.Find(123)
	})

	service.ExpectReq("POST", "/users").WithBody("name", "User McUser").WithBody("email", "user@example.com").Do(func() {
		client.Create(User{
			Name:  "User McUser",
			Email: "user@example.com",
		})
	})

	fmt.Println(service.Summary())
	// Output:
	// Expectations
	// 	GET /users/123 - passed
	// 	POST /users with body matching name="User McUser"body matching email="user@example.com" - passed
}
