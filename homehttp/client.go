package homehttp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const (
	defaultUserAgent   = "homehttp.Client"
	defaultTimeout     = 30 * time.Second
	defaultRetries     = 0
	defaultBackoffTime = 300 * time.Millisecond

	respSizeLimit = int64(10 * 1024 * 1024) // 10MB
)

var ErrorTimeout = errors.New("request timeout")

// Client is a wrapper for default http.Client.
type Client struct {
	baseClient *http.Client
	logger     *zerolog.Logger
	retryer    RetryStrategy

	backoff      BackoffStrategy
	retryWaitMin time.Duration
	retryWaitMax time.Duration

	maxRetries int
}

// NewClient returns a new Client.
func NewClient(opts ...ClientOption) *Client {
	defaultLogger := zerolog.Nop()

	cfg := &clientConfig{
		AppName: defaultUserAgent,
		Timeout: defaultTimeout,
		Logger:  &defaultLogger,

		Retryer:    NoRetry,
		MaxRetries: defaultRetries,

		Backoff: ConstantBackoff(defaultBackoffTime),
	}

	for _, o := range opts {
		o.apply(cfg)
	}

	return buildClient(cfg)
}

type clientConfig struct {
	AppName              string
	Timeout              time.Duration
	TransportMiddlewares []roundTripperMiddleware
	Headers              map[string]string

	Retryer    RetryStrategy
	MaxRetries int

	Backoff      BackoffStrategy
	MinRetryWait time.Duration
	MaxRetryWait time.Duration

	Logger *zerolog.Logger
}

func buildClient(cfg *clientConfig) *Client {
	cfg.TransportMiddlewares = append(cfg.TransportMiddlewares, clientUserAgent(cfg.AppName))

	return &Client{
		baseClient: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: chainRoundTrippers(http.DefaultTransport, cfg.TransportMiddlewares...),
		},
		logger:     cfg.Logger,
		retryer:    cfg.Retryer,
		backoff:    cfg.Backoff,
		maxRetries: cfg.MaxRetries,
	}
}

// DoJSON executes a request.
func (c *Client) DoJSON(ctx context.Context, method, url string, payload any) (*http.Response, error) {
	req, err := NewRequestJSON(ctx, method, url, payload)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	var (
		reqBodyBytes []byte
		resp         *http.Response
		shouldRetry  bool
		doErr        error
	)

	if req.Body != nil {
		reqBodyBytes, _ = io.ReadAll(req.Body)
	}

	for i := 0; ; i++ {
		if reqBodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
		}

		resp, doErr = c.baseClient.Do(req)
		shouldRetry = c.retryer.Classify(req.Context(), resp, doErr)

		if doErr != nil {
			c.logger.Debug().Err(err).
				Str("method", req.Method).
				Str("url", req.URL.String()).
				Msg("failed to execute request")
		}

		if !shouldRetry {
			break
		}

		// We do this before drainBody because there's no need for the I/O if
		// we're breaking out
		remainAtt := c.maxRetries - i
		if remainAtt <= 0 {
			break
		}

		// We're going to retry, consume any response to reuse the connection.
		if doErr == nil {
			c.drainBody(resp.Body)
		}

		wait := c.backoff.Backoff(c.retryWaitMin, c.retryWaitMax, i, resp)

		// Wait before retrying
		timer := time.NewTimer(wait)
		select {
		case <-req.Context().Done():
			timer.Stop()
			c.baseClient.CloseIdleConnections()

			return nil, req.Context().Err()
		case <-timer.C:
		}

	}

	if doErr == nil && !shouldRetry {
		return resp, nil
	}

	// retry was not successful
	return nil, ErrorResponse{Response: resp, Original: doErr}
}

func (c *Client) drainBody(body io.ReadCloser) {
	if body != nil {
		_, _ = io.Copy(io.Discard, io.LimitReader(body, respSizeLimit))
		_ = body.Close()
	}
}

type ErrorResponse struct {
	Response *http.Response
	Original error
}

func (r ErrorResponse) Error() string {
	if r.Response == nil {
		return r.Original.Error()
	}

	return fmt.Sprintf("%v %v: %d",
		r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode,
	)
}
