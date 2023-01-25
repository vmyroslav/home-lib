package homehttp

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const defaultClientName = "homehttp.Client"

// Client is a wrapper for default http.Client.
type Client struct {
	*http.Client
}

// NewClient returns a new Client.
func NewClient(opts ...ClientOption) *Client {
	defaultLogger := zerolog.Nop()

	cfg := &clientConfig{
		AppName: defaultClientName,
		Timeout: 30 * time.Second,
		Logger:  &defaultLogger,
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

	Logger *zerolog.Logger
}

func buildClient(cfg *clientConfig) *Client {
	cfg.TransportMiddlewares = append(cfg.TransportMiddlewares, clientUserAgent(cfg.AppName))

	return &Client{
		Client: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: chainRoundTrippers(http.DefaultTransport, cfg.TransportMiddlewares...),
		},
	}
}

// Do executes a request.
func (c *Client) Do(ctx context.Context, method, url string, payload any) (*http.Response, error) {
	req, err := NewRequest(ctx, method, url, payload)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		return nil, errors.Wrap(err, "failed to execute request")
	}

	err = checkResponse(resp)

	return resp, err
}

type ErrorResponse struct {
	Response *http.Response
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d",
		r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode,
	)
}
func checkResponse(resp *http.Response) error {
	if http.StatusOK <= resp.StatusCode && resp.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	return &ErrorResponse{Response: resp}
}
