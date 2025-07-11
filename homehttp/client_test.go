package homehttp

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestClientDo(t *testing.T) {
	t.Parallel()

	// Create a test server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/success":
			w.WriteHeader(http.StatusOK)
		case "/created":
			w.WriteHeader(http.StatusCreated)
		case "/badRequest":
			w.WriteHeader(http.StatusBadRequest)
		case "/unauthorized":
			w.WriteHeader(http.StatusUnauthorized)
		case "/redirect":
			w.WriteHeader(http.StatusTemporaryRedirect)
		case "/failure":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	defer testServer.Close()

	// Define the test cases
	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
	}{
		{
			name:           "GET success",
			method:         "GET",
			url:            testServer.URL + "/success",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST success",
			method:         "POST",
			url:            testServer.URL + "/created",
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "GET bad request",
			method:         "GET",
			url:            testServer.URL + "/badRequest",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "GET unauthorized",
			method:         "GET",
			url:            testServer.URL + "/unauthorized",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "GET redirect",
			method:         "GET",
			url:            testServer.URL + "/redirect",
			expectedStatus: http.StatusTemporaryRedirect,
		},
		{
			name:           "GET failure",
			method:         "GET",
			url:            testServer.URL + "/failure",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "GET not found",
			method:         "GET",
			url:            testServer.URL + "/notfound",
			expectedStatus: http.StatusNotFound,
		},
	}

	// Create a new client
	client := NewClient()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.DoJSON(context.Background(), tt.method, tt.url, nil)
			if err != nil {
				var errResp *ResponseError
				if !errors.As(err, &errResp) {
					t.Fatal(err, "unexpected error type received")
				}

				assert.Equal(
					t,
					tt.expectedStatus,
					err.(*ResponseError).Response.StatusCode,
					"status code should match",
				)
			}

			assert.NotNil(t, resp, "response should not be nil")
			assert.Equal(t, tt.expectedStatus, resp.StatusCode, "status code should match")
		})
	}
}

// TestClientDoWithTimeoutClientOption tests that the client will time out if the request takes longer than the timeout.
func TestClientDoWithTimeoutClientOption(t *testing.T) {
	t.Parallel()

	timeout := 100 * time.Millisecond

	// Create a test server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/success":
			w.WriteHeader(http.StatusOK)
		case "/timeout":
			time.Sleep(timeout + time.Millisecond*10)
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	t.Cleanup(testServer.Close)

	// Define the test cases
	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
		expectTimeout  bool
	}{
		{
			name:           "GET success",
			method:         "GET",
			url:            testServer.URL + "/success",
			expectedStatus: http.StatusOK,
			expectTimeout:  false,
		},
		{
			name:           "GET timeout",
			method:         "GET",
			url:            testServer.URL + "/timeout",
			expectedStatus: http.StatusInternalServerError,
			expectTimeout:  true,
		},
	}

	// Create a new client with a timeout
	client := NewClient(WithTimeout(timeout))

	// Run the tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resp, err := client.DoJSON(context.Background(), tt.method, tt.url, nil)
			if err != nil {
				if !tt.expectTimeout {
					t.Fatal(err, "unexpected error type received")
				}

				var urlErr *url.Error
				if errors.As(err, &urlErr) {
					assert.True(t, urlErr.Timeout(), "timeout should be true")
				}

				assert.Nil(t, resp, "response should be nil")

				return
			}

			assert.NotNil(t, resp, "response should not be nil")
			assert.Equal(t, tt.expectedStatus, resp.StatusCode, "status code should match")
		})
	}
}

// TestClientDoWithContextTimeout tests that the client will time out
// if the request takes longer than the context timeout.
func TestClientDoWithContextTimeout(t *testing.T) {
	t.Parallel()

	timeout := 100 * time.Millisecond

	// Create a test server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/success":
			w.WriteHeader(http.StatusOK)
		case "/timeout":
			time.Sleep(timeout + time.Millisecond*10)
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	t.Cleanup(testServer.Close)

	// Define the test cases
	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
		expectTimeout  bool
	}{
		{
			name:           "GET success",
			method:         "GET",
			url:            testServer.URL + "/success",
			expectedStatus: http.StatusOK,
			expectTimeout:  false,
		},
		{
			name:           "GET timeout",
			method:         "GET",
			url:            testServer.URL + "/timeout",
			expectedStatus: http.StatusInternalServerError,
			expectTimeout:  true,
		},
	}

	// Create a new client
	client := NewClient()

	// Run the tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			resp, err := client.DoJSON(ctx, tt.method, tt.url, nil)
			if err != nil {
				if !tt.expectTimeout {
					t.Fatal(err, "unexpected error type received")
				}

				var expErr ResponseError

				assert.ErrorAs(t, err, &expErr, "expected ResponseError")

				if respErr, ok := err.(ResponseError); ok && respErr.Response != nil {
					require.ErrorIs(t, respErr.Original, context.DeadlineExceeded, "context deadline exceeded error should be returned")
				}

				return
			}

			assert.NotNil(t, resp, "response should not be nil")
			assert.Equal(t, tt.expectedStatus, resp.StatusCode, "status code should match")
		})
	}
}

// TestClientDoWithAppNameOption tests that the client will set the User-Agent header to the app name.
func TestClientDoWithAppNameOption(t *testing.T) {
	t.Parallel()

	// Create a test server
	testServer := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test", r.Header.Get("User-Agent"))
	}))
	defer testServer.Close()

	// Create a new client
	client := NewClient(WithAppName("test"))

	// Run the tests
	resp, err := client.DoJSON(context.Background(), "GET", testServer.URL, nil)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestClientDoWithBasicAuthOption(t *testing.T) {
	t.Parallel()

	// Create a test server
	testServer := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("username:password"))
		assert.Equal(t, expectedAuth, auth)
	}))
	defer testServer.Close()

	// Create a new client with basic auth option
	client := NewClient(WithBasicAuth("username", "password"))

	// Run the test
	resp, err := client.DoJSON(context.Background(), "GET", testServer.URL, nil)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestClientDoWithHeaderOption(t *testing.T) {
	t.Parallel()

	// Create a test server
	testServer := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test", r.Header.Get("key"))
	}))
	defer testServer.Close()

	// Create a new client
	client := NewClient(WithHeader("key", "test"))

	// Run the tests
	resp, err := client.DoJSON(context.Background(), "GET", testServer.URL, nil)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestClientDoWithAuthorizationTokenOption tests that the client will set the Authorization header to the token.
func TestClientDoWithAuthorizationTokenOption(t *testing.T) {
	t.Parallel()

	// Create a test server
	testServer := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		expectedAuth := "Bearer " + "token"
		assert.Equal(t, expectedAuth, auth)
	}))
	defer testServer.Close()

	// Create a new client with authorization token option
	client := NewClient(WithAuthorizationToken(Token{
		AccessToken: "token",
		ExpiresAt:   time.Now().Add(time.Hour),
		Type:        "Bearer",
	}))

	// Run the test
	resp, err := client.DoJSON(context.Background(), "GET", testServer.URL, nil)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestBasicAuthTokenValidation verifies that basic auth tokens don't expire
func TestBasicAuthTokenValidation(t *testing.T) {
	t.Parallel()

	// Create basic auth token provider
	tokenProvider := basicAuthorization("user", "pass")

	// Get token
	token, err := tokenProvider.GetToken(context.Background())
	require.NoError(t, err)

	// Verify token properties
	assert.Equal(t, "Basic", token.Type)
	assert.NotEmpty(t, token.AccessToken)
	assert.True(t, token.ExpiresAt.IsZero(), "Basic auth token should not have expiration")
	assert.True(t, token.IsValid(), "Basic auth token should be valid")

	// Token should still be valid after some time
	time.Sleep(10 * time.Millisecond)
	assert.True(t, token.IsValid(), "Basic auth token should remain valid")
}

// TestTokenIsValidWithExpiration verifies IsValid works correctly with expiring tokens
func TestTokenIsValidWithExpiration(t *testing.T) {
	t.Parallel()

	// Test valid token with future expiration
	validToken := Token{
		AccessToken: "valid-token",
		ExpiresAt:   time.Now().Add(time.Hour),
		Type:        "Bearer",
	}
	assert.True(t, validToken.IsValid())

	// Test expired token
	expiredToken := Token{
		AccessToken: "expired-token",
		ExpiresAt:   time.Now().Add(-time.Hour),
		Type:        "Bearer",
	}
	assert.False(t, expiredToken.IsValid())

	// Test token with no access token
	emptyToken := Token{
		AccessToken: "",
		ExpiresAt:   time.Now().Add(time.Hour),
		Type:        "Bearer",
	}
	assert.False(t, emptyToken.IsValid())

	// Test token with zero expiration (like basic auth)
	noExpirationToken := Token{
		AccessToken: "no-expiry-token",
		ExpiresAt:   time.Time{}, // Zero time
		Type:        "Basic",
	}
	assert.True(t, noExpirationToken.IsValid())
}
