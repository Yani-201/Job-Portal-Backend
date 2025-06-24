package testutils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// RequestOptions contains options for making HTTP requests in tests
type RequestOptions struct {
	Method string
	URL    string
	Body   interface{}
	Token  string
}

// MakeRequest makes an HTTP request for testing
func MakeRequest(t *testing.T, router http.Handler, opts RequestOptions) *http.Response {
	t.Helper()

	var body io.Reader
	if opts.Body != nil {
		jsonData, err := json.Marshal(opts.Body)
		require.NoError(t, err)
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(opts.Method, opts.URL, body)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	if opts.Token != "" {
		req.Header.Set("Authorization", "Bearer "+opts.Token)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	return rr.Result()
}

// ParseJSONResponse parses the JSON response body into the target struct
func ParseJSONResponse(t *testing.T, resp *http.Response, target interface{}) {
	t.Helper()

	defer resp.Body.Close()
	err := json.NewDecoder(resp.Body).Decode(target)
	require.NoError(t, err)
}
