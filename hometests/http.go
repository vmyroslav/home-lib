package hometests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// HTTPServer creates a test HTTP server with the provided handler.
// Returns the server and its base URL.
func HTTPServer(t *testing.T, handler http.Handler) (*httptest.Server, string) {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return server, server.URL
}

// HTTPSServer creates a test HTTPS server with the provided handler.
// Returns the server and its base URL.
func HTTPSServer(t *testing.T, handler http.Handler) (*httptest.Server, string) {
	t.Helper()

	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)

	return server, server.URL
}

// JSONServer creates a test server that responds with JSON.
func JSONServer(t *testing.T, statusCode int, response any) (*httptest.Server, string) {
	t.Helper()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		if response != nil {
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Errorf("failed to encode JSON response: %v", err)
			}
		}
	})

	return HTTPServer(t, handler)
}

// ErrorServer creates a test server that always returns an error status.
func ErrorServer(t *testing.T, statusCode int, message string) (*httptest.Server, string) {
	t.Helper()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, message, statusCode)
	})

	return HTTPServer(t, handler)
}

// EchoServer creates a test server that echoes the request body.
func EchoServer(t *testing.T) (*httptest.Server, string) {
	t.Helper()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))

		if _, err := io.Copy(w, r.Body); err != nil {
			t.Errorf("failed to copy request body: %v", err)
		}
	})

	return HTTPServer(t, handler)
}

// RequestCapture captures HTTP requests for inspection in tests.
type RequestCapture struct {
	Requests []*http.Request
	Bodies   []string
}

// CaptureServer creates a test server that captures all requests.
func CaptureServer(t *testing.T, statusCode int, response string) (*httptest.Server, string, *RequestCapture) {
	t.Helper()

	capture := &RequestCapture{}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture the request
		capture.Requests = append(capture.Requests, r)

		// Capture the body
		if r.Body != nil {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Errorf("failed to read request body: %v", err)
			}

			capture.Bodies = append(capture.Bodies, string(body))
			// Reset body for handler
			r.Body = io.NopCloser(strings.NewReader(string(body)))
		}

		w.WriteHeader(statusCode)

		if _, err := w.Write([]byte(response)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	})

	server, url := HTTPServer(t, handler)

	return server, url, capture
}

// MockRoundTripper is a mock http.RoundTripper for testing HTTP clients.
type MockRoundTripper struct {
	Response *http.Response
	Error    error
	Requests []*http.Request
}

// RoundTrip implements http.RoundTripper.
func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	m.Requests = append(m.Requests, req)

	if m.Error != nil {
		return nil, m.Error
	}

	if m.Response == nil {
		return &http.Response{
			StatusCode: 200,
			Status:     "200 OK",
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("")),
			Request:    req,
		}, nil
	}

	m.Response.Request = req

	return m.Response, nil
}

// NewMockResponse creates a mock HTTP response for testing.
func NewMockResponse(statusCode int, body string, headers map[string]string) *http.Response {
	resp := &http.Response{
		StatusCode: statusCode,
		Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	for key, value := range headers {
		resp.Header.Set(key, value)
	}

	return resp
}
