package homehttp

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRequest_WithBody(t *testing.T) {
	ctx := context.Background()
	body := map[string]string{"key": "value"}

	req, err := NewRequestJSON(ctx, http.MethodPost, "http://localhost", body)

	assert.NoError(t, err)
	assert.NotNil(t, req)
	assert.Equal(t, http.MethodPost, req.Method)
	assert.Equal(t, "http://localhost", req.URL.String())
	assert.Equal(t, defaultContentType, req.Header.Get("Content-Type"))

	buf := new(bytes.Buffer)
	buf.ReadFrom(req.Body)
	actual := strings.TrimSuffix(buf.String(), "\n") // trim trailing newline

	assert.Equal(t, `{"key":"value"}`, actual)
}

func TestNewRequest_WithoutBody(t *testing.T) {
	ctx := context.Background()

	req, err := NewRequestJSON(ctx, http.MethodGet, "http://localhost", nil)

	assert.NoError(t, err)
	assert.NotNil(t, req)
	assert.Equal(t, http.MethodGet, req.Method)
	assert.Equal(t, "http://localhost", req.URL.String())
	assert.Equal(t, "", req.Header.Get("Content-Type"))
}

func TestNewRequest_InvalidURL(t *testing.T) {
	ctx := context.Background()
	body := map[string]string{"key": "value"}

	req, err := NewRequestJSON(ctx, http.MethodPost, ":", body)

	assert.Error(t, err)
	assert.Nil(t, req)
}

func TestNewRequest_InvalidBody(t *testing.T) {
	ctx := context.Background()
	body := make(chan int)

	req, err := NewRequestJSON(ctx, http.MethodPost, "http://localhost", body)

	assert.Error(t, err)
	assert.Nil(t, req)
}
