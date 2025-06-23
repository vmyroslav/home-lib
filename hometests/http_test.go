package hometests

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestHTTPServer(t *testing.T) {
	t.Parallel()

	t.Run("creates test server", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test response"))
		})

		server, url := HTTPServer(t, handler)

		// Make request to server
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}

		if string(body) != "test response" {
			t.Errorf("Expected 'test response', got %s", string(body))
		}

		// Verify server is cleaned up automatically
		server.Close()
	})
}

func TestJSONServer(t *testing.T) {
	t.Parallel()

	t.Run("responds with JSON", func(t *testing.T) {
		t.Parallel()

		response := map[string]string{"message": "hello"}
		_, url := JSONServer(t, http.StatusOK, response)

		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected JSON content type, got %s", resp.Header.Get("Content-Type"))
		}

		var result map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode JSON: %v", err)
		}

		if result["message"] != "hello" {
			t.Errorf("Expected message 'hello', got %s", result["message"])
		}
	})
}

func TestErrorServer(t *testing.T) {
	t.Parallel()

	t.Run("responds with error", func(t *testing.T) {
		t.Parallel()

		message := "something went wrong"
		_, url := ErrorServer(t, http.StatusInternalServerError, message)

		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}

		if !strings.Contains(string(body), message) {
			t.Errorf("Expected body to contain %q, got %q", message, string(body))
		}
	})
}

func TestEchoServer(t *testing.T) {
	t.Parallel()

	t.Run("echoes request body", func(t *testing.T) {
		t.Parallel()

		_, url := EchoServer(t)

		requestBody := "test request body"

		resp, err := http.Post(url, "text/plain", strings.NewReader(requestBody))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}

		if string(body) != requestBody {
			t.Errorf("Expected echo %q, got %q", requestBody, string(body))
		}
	})
}

func TestCaptureServer(t *testing.T) {
	t.Parallel()

	t.Run("captures requests", func(t *testing.T) {
		t.Parallel()

		_, url, capture := CaptureServer(t, http.StatusOK, "response")

		// Make a few requests
		http.Get(url + "/path1")
		http.Post(url+"/path2", "text/plain", strings.NewReader("body1"))

		// Make PUT request manually since http.Put doesn't exist
		req, _ := http.NewRequest("PUT", url+"/path3", strings.NewReader("body2"))
		req.Header.Set("Content-Type", "application/json")
		http.DefaultClient.Do(req)

		// Verify requests were captured
		if len(capture.Requests) != 3 {
			t.Errorf("Expected 3 captured requests, got %d", len(capture.Requests))
		}

		if capture.Requests[0].Method != "GET" {
			t.Errorf("Expected first request to be GET, got %s", capture.Requests[0].Method)
		}

		if capture.Requests[1].Method != "POST" {
			t.Errorf("Expected second request to be POST, got %s", capture.Requests[1].Method)
		}

		if capture.Requests[2].Method != "PUT" {
			t.Errorf("Expected third request to be PUT, got %s", capture.Requests[2].Method)
		}

		if len(capture.Bodies) != 3 {
			t.Errorf("Expected 3 captured bodies, got %d", len(capture.Bodies))
		}

		if capture.Bodies[1] != "body1" {
			t.Errorf("Expected second body to be 'body1', got %s", capture.Bodies[1])
		}
	})
}

func TestMockRoundTripper(t *testing.T) {
	t.Parallel()

	t.Run("mocks HTTP round trips", func(t *testing.T) {
		t.Parallel()

		mockResponse := NewMockResponse(200, "test response", map[string]string{
			"Content-Type": "text/plain",
		})

		mock := &MockRoundTripper{
			Response: mockResponse,
		}

		client := &http.Client{
			Transport: mock,
		}

		resp, err := client.Get("http://example.com")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		if len(mock.Requests) != 1 {
			t.Errorf("Expected 1 captured request, got %d", len(mock.Requests))
		}

		if mock.Requests[0].URL.String() != "http://example.com" {
			t.Errorf("Expected URL http://example.com, got %s", mock.Requests[0].URL.String())
		}
	})
}
