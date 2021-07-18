package hex_test

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"

	"github.com/meagar/hex"
)

func ExampleExpectation_RespondWith() {
	server := hex.NewServer(&testing.T{}, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// The default behavior of the mock service
		fmt.Fprintf(rw, "ok")
	}))

	// Override handler for any requests that match this expectation
	server.ExpectReq("GET", "/foo").Once().RespondWith(200, "mock response")

	if resp, err := http.Get(server.URL + "/foo"); err != nil {
		panic(err)
	} else {
		// Verify the mock response was received
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)

		if string(body) != "mock response" {
			log.Panicf(`Expected body to match "mock response", got %s (%v)`, body, err)
		}
	}

	// Should hit the server's default handler
	server.ExpectReq("GET", "/bar")
	if resp, err := http.Get(server.URL + "/bar"); err != nil {
		log.Panicf("Unexpected error contacting mock service: %v", err)
	} else {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)

		if string(body) != "ok" {
			log.Panicf(`Expected body to match "ok", got %s (%v)`, body, err)
		}
	}

	fmt.Println(server.Summary())
	// Output:
	// Expectations
	// 	GET /foo - passed
	// 	GET /bar - passed
}

func TestExpectationRespondWith(t *testing.T) {
	t.Run("A mock response prevents the original handler from being called when AndCallThrough is not used", func(t *testing.T) {
		server := hex.NewServer(t, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			io.WriteString(rw, "ok from server")
		}))

		server.ExpectReq("GET", "/foo").RespondWith(401, "ok from exp")

		resp, err := http.Get(server.URL + "/foo")
		if err != nil {
			panic(err)
		}
		if resp.StatusCode != 401 {
			t.Errorf("Expected response status to be 401, got %d", resp.StatusCode)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if string(body) != "ok from exp" {
			t.Errorf(`Expected response body to be "ok from exp", got %q`, string(body))
		}
	})

	t.Run("A mock response calls through to the original handler when AndCallThrough is used", func(t *testing.T) {
		server := hex.NewServer(t, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			io.WriteString(rw, "ok from server")
		}))

		server.ExpectReq("GET", "/foo").RespondWith(401, "ok from exp").AndCallThrough()

		resp, err := http.Get(server.URL + "/foo")
		if err != nil {
			panic(err)
		}
		if resp.StatusCode != 401 {
			t.Errorf("Expected response status to be 401, got %d", resp.StatusCode)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if string(body) != "ok from expok from server" {
			t.Errorf(`Expected response body to be "ok from exp", got %q`, string(body))
		}
	})
}
