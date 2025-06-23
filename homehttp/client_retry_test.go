package homehttp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientDoWithRetry(t *testing.T) {
	// Define the test cases
	tests := []struct {
		name                 string
		method               string
		path                 string
		successStatusCode    int
		failureStatusCode    int
		successResponseBody  string
		failureResponseBody  string
		failuresNum          int
		maxRetries           int
		serverTimeout        time.Duration
		clientTimeout        time.Duration
		expectedCalls        int
		expectedRequestBody  string
		expectedResponseBody string
		expectedStatusCode   int
		expectFailure        bool
	}{
		{
			name:                 "GET Success without retry",
			method:               http.MethodGet,
			path:                 "/success",
			successStatusCode:    http.StatusOK,
			failureStatusCode:    http.StatusInternalServerError,
			successResponseBody:  `{"message":"hello"}`,
			failureResponseBody:  "",
			failuresNum:          0,
			maxRetries:           3,
			expectedCalls:        1,
			expectedRequestBody:  "",
			expectedResponseBody: `{"message":"hello"}`,
			expectedStatusCode:   http.StatusOK,
		},
		{
			name:                 "GET Success with retry",
			method:               http.MethodGet,
			path:                 "/success",
			successStatusCode:    http.StatusOK,
			failureStatusCode:    http.StatusInternalServerError,
			successResponseBody:  `{"message":"hello"}`,
			failureResponseBody:  `{"error":"Internal Server Error"}`,
			failuresNum:          2,
			maxRetries:           3,
			expectedCalls:        3,
			expectedRequestBody:  "",
			expectedResponseBody: `{"message":"hello"}`,
			expectedStatusCode:   http.StatusOK,
		},
		{
			name:                 "GET failed after retries",
			method:               http.MethodGet,
			path:                 "/success",
			failureStatusCode:    http.StatusInternalServerError,
			successStatusCode:    http.StatusOK,
			successResponseBody:  `{}`,
			failureResponseBody:  `{}`,
			failuresNum:          5,
			maxRetries:           3,
			expectedCalls:        4,
			expectedRequestBody:  "{}",
			expectedResponseBody: `{}`,
			expectedStatusCode:   http.StatusInternalServerError,
			expectFailure:        true,
		},
		{
			name:                 "POST success after retries",
			method:               http.MethodPost,
			path:                 "/success",
			successStatusCode:    http.StatusCreated,
			failureStatusCode:    http.StatusInternalServerError,
			successResponseBody:  `{"id":1, "name":"test"}`,
			failureResponseBody:  `{"error":"Internal Server Error"}`,
			failuresNum:          3,
			maxRetries:           3,
			expectedCalls:        4,
			expectedRequestBody:  `{"name":"test"}`,
			expectedResponseBody: `{"id":1, "name":"test"}`,
			expectedStatusCode:   http.StatusCreated,
		},
		{
			name:                 "POST Error with retry",
			method:               "POST",
			path:                 "/error",
			failureStatusCode:    http.StatusInternalServerError,
			successStatusCode:    http.StatusCreated,
			successResponseBody:  `{"id":1, "name":"test"}`,
			failureResponseBody:  `{"error":"Internal Server Error"}`,
			failuresNum:          2,
			maxRetries:           3,
			expectedCalls:        3,
			expectedRequestBody:  `{"name":"test"}`,
			expectedResponseBody: `{"id":1, "name":"test"}`,
			expectedStatusCode:   http.StatusCreated,
		},
		{
			name:                 "GET failed after first server timeout without retry",
			method:               http.MethodGet,
			path:                 "/timeout",
			successStatusCode:    http.StatusOK,
			failureStatusCode:    http.StatusBadGateway,
			successResponseBody:  "",
			failureResponseBody:  "",
			failuresNum:          5,
			maxRetries:           3,
			serverTimeout:        100 * time.Millisecond,
			clientTimeout:        50 * time.Millisecond,
			expectedCalls:        1,
			expectedRequestBody:  "",
			expectedResponseBody: "",
			expectedStatusCode:   http.StatusBadGateway,
			expectFailure:        true,
		},
		{
			name:                 "GET client errors should not be retried",
			method:               http.MethodGet,
			path:                 "/client-error",
			successStatusCode:    http.StatusOK,
			failureStatusCode:    http.StatusBadRequest,
			successResponseBody:  `{"message":"hello"}`,
			failureResponseBody:  `{"error":"Bad Request"}`,
			failuresNum:          1,
			maxRetries:           3,
			expectedCalls:        1,
			expectedRequestBody:  "",
			expectedResponseBody: `{"error":"Bad Request"}`,
			expectedStatusCode:   http.StatusBadRequest,
		},
		{
			name:                 "GET malformed response",
			method:               http.MethodGet,
			path:                 "/malformed",
			successStatusCode:    http.StatusOK,
			failureStatusCode:    http.StatusOK,
			successResponseBody:  `{"message":"hello"}`,
			failureResponseBody:  `{"error":`,
			failuresNum:          1,
			maxRetries:           3,
			expectedCalls:        1,
			expectedRequestBody:  "",
			expectedResponseBody: `{"error":`,
			expectedStatusCode:   http.StatusOK,
		},
	}

	// Run the tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new client with a retry strategy
			client := NewClient(
				WithRetryStrategy(RetryOn500x),
				WithMaxRetries(tt.maxRetries),
				WithTimeout(tt.clientTimeout),
			)

			var callCount int

			// Create a test server
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++

				// check if the request body is not corrupted
				actualPayload, err := io.ReadAll(r.Body)
				assert.NoError(t, err)

				var actualRequest string
				err = json.Unmarshal(actualPayload, &actualRequest)
				assert.NoError(t, err)

				assert.Equal(t, tt.expectedRequestBody, actualRequest, "request body should match")

				// sleep to simulate server timeout
				if tt.path == "/timeout" {
					time.Sleep(tt.serverTimeout)
				}

				if callCount <= tt.failuresNum {
					w.WriteHeader(tt.failureStatusCode)
					_, _ = w.Write([]byte(tt.failureResponseBody))

					return
				}

				w.WriteHeader(tt.successStatusCode)
				_, _ = w.Write([]byte(tt.successResponseBody))
			}))
			defer testServer.Close()

			resp, err := client.DoJSON(context.Background(), tt.method, testServer.URL+tt.path, tt.expectedRequestBody)
			if err != nil && resp != nil {
				defer resp.Body.Close()
			}

			if tt.expectFailure { //nolint:wsl
				var expErr ResponseError

				require.Error(t, err, "expected error, but got nil")
				require.ErrorAs(t, err, &expErr, "expected ResponseError")
				assert.Nil(t, resp, "response should be nil")
				assert.Equal(t, tt.expectedCalls, callCount, "number of calls should match")

				if respErr, ok := err.(ResponseError); ok && respErr.Response != nil {
					bodyBytes, _ := io.ReadAll(err.(ResponseError).Response.Body)
					assert.Equal(t, tt.expectedResponseBody, string(bodyBytes), "response body should match")
				}

				return
			}

			require.NoError(t, err, "expected no error, but got %v", err)
			assert.NotNil(t, resp, "response should not be nil")
			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode, "status code should match")
			assert.Equal(t, tt.expectedCalls, callCount, "number of calls should match")

			// Check if the response body is not corrupted
			bodyBytes, _ := io.ReadAll(resp.Body)
			assert.Equal(t, tt.expectedResponseBody, string(bodyBytes), "response body should match")
		})
	}
}

// TestClientRetryBackoffTiming verifies that retry wait times are properly configured and used
func TestClientRetryBackoffTiming(t *testing.T) {
	t.Parallel()

	var requestTimes []time.Time

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestTimes = append(requestTimes, time.Now())

		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("server error"))
	}))
	defer server.Close()

	// create client with specific retry configuration
	minWait := 50 * time.Millisecond
	maxWait := 200 * time.Millisecond
	client := NewClient(
		WithRetryStrategy(RetryOn500x),
		WithMaxRetries(2), // Will make 3 requests total (initial + 2 retries)
		WithRetryWaitTimes(minWait, maxWait),
		WithBackoffStrategy(LinearBackoff(25*time.Millisecond)), // Incremental backoff
	)

	start := time.Now()
	_, err := client.DoJSON(context.Background(), "GET", server.URL, nil)
	duration := time.Since(start)

	require.Error(t, err)

	assert.Len(t, requestTimes, 3, "Expected 3 requests (initial + 2 retries)")

	// Verify minimum delays between requests
	if len(requestTimes) >= 2 {
		firstDelay := requestTimes[1].Sub(requestTimes[0])
		assert.GreaterOrEqual(t, firstDelay, minWait, "First retry delay should be at least minWait")
	}

	if len(requestTimes) >= 3 {
		secondDelay := requestTimes[2].Sub(requestTimes[1])
		assert.GreaterOrEqual(t, secondDelay, minWait, "Second retry delay should be at least minWait")
	}

	// Total duration should include retry delays
	expectedMinDuration := 2 * minWait // 2 retry delays
	assert.GreaterOrEqual(t, duration, expectedMinDuration, "Total duration should include retry delays")
}

// TestClientRetryWaitCustom verifies that custom retry wait times are properly set
func TestClientRetryWaitCustom(t *testing.T) {
	t.Parallel()

	customMin := 250 * time.Millisecond
	customMax := 5 * time.Second

	client := NewClient(
		WithRetryWaitTimes(customMin, customMax),
	)

	assert.Equal(t, customMin, client.retryWaitMin, "retryWaitMin should match custom value")
	assert.Equal(t, customMax, client.retryWaitMax, "retryWaitMax should match custom value")
}
