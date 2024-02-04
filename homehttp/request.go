package homehttp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

const defaultContentType = "application/json"

// NewRequestJSON creates a new http request.
func NewRequestJSON(ctx context.Context, method, urlStr string, body any) (*http.Request, error) {
	var (
		buf     io.ReadWriter
		headers = map[string]string{}
	)

	if body != nil {
		buf = new(bytes.Buffer)
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)

		if err := enc.Encode(body); err != nil {
			return nil, errors.Wrap(err, "failed to encode body")
		}

		headers["Content-Type"] = defaultContentType
	}

	req, err := http.NewRequestWithContext(ctx, method, urlStr, buf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return req, nil
}
