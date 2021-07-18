package hex_test

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/meagar/hex"
)

func TestServer(t *testing.T) {

	t.Run("NewServer return a new server which accepts non-SSL requests", func(t *testing.T) {
		server := hex.NewServer(t, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			fmt.Fprintf(rw, "OK")
		}))

		if !strings.HasPrefix(server.URL, "http://") {
			t.Errorf("Expected server URL to start with http://, got %q", server.URL)
		}

		// A request from any HTTP client should work
		resp, err := http.Get(server.URL + "/foo")

		if err != nil {
			t.Errorf("Expected GET to a non-TLS server to work, got %v", err)
		} else {
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if string(body) != "OK" || err != nil {
				t.Errorf("Expected server to respond with OK, got %s/%v", body, err)
			}
		}
	})

	t.Run("NewTLSServer return a new server which accepts SSL requests", func(t *testing.T) {
		server := hex.NewTLSServer(t, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			fmt.Fprintf(rw, "OK")
		}))

		if !strings.HasPrefix(server.URL, "https://") {
			t.Errorf("Expected server URL to start with https://, got %q", server.URL)
		}

		// This call should fail, the "real" http default client should not accept the mock server's certificate
		_, err := http.Get(server.URL + "/foo")
		if err == nil {
			t.Errorf("GET to a mock TLS service succeeded when it should have failed due to an untrusted cert")
		}

		// This call should succeed, the server's client is pre-configured to accept the server's certificate
		resp, err := server.Client().Get(server.URL + "/foo")
		if err != nil {
			t.Errorf("GET to a mock TLS service failed when it should have passed: %v", err)
		} else {
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if string(body) != "OK" || err != nil {
				t.Errorf("Expected server to respond with OK, got %s/%v", body, err)
			}
		}
	})
}
