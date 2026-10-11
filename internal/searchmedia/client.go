package searchmedia

import (
	"io"
	"net/http"
	"time"
)

type HTTPResponse struct {
	// 0 means no response: wrong host, connection refused, or timeout.
	StatusCode int
	Body       []byte
}

func (r HTTPResponse) IsOK() bool {
	return r.StatusCode == http.StatusOK
}

type HTTPClient interface {
	Get(url string, headers map[string]string) HTTPResponse
}

type StdHTTPClient struct {
	client *http.Client
}

// The timeout covers the whole request, body included, so a slow-dripping answer cannot hold it longer.
func NewHTTPClient(timeout time.Duration) *StdHTTPClient {
	return &StdHTTPClient{client: &http.Client{Timeout: timeout}}
}

func (c *StdHTTPClient) Get(url string, headers map[string]string) HTTPResponse {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return HTTPResponse{}
	}

	for name, value := range headers {
		request.Header.Set(name, value)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return HTTPResponse{}
	}
	defer response.Body.Close()

	// Don't pass error responses to the mapper.
	if response.StatusCode >= 400 {
		return HTTPResponse{StatusCode: response.StatusCode}
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return HTTPResponse{}
	}

	return HTTPResponse{StatusCode: response.StatusCode, Body: body}
}
