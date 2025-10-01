package httpclient

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type Client struct {
	client  *http.Client
	limiter *rate.Limiter
}

// Creates new Client instanse.
//
// Params:
//   - rpm: how many request per minute (maximum requests limit per minute).
//   - burst: how many requests fire at once. Set to '1' for smooth working.
//   - timeout: maximum time for each request.
func New(rpm int, burst int, timeout time.Duration) *Client {
	rps := rate.Limit(float64(rpm) / 60.0)

	return &Client{
		client:  &http.Client{Timeout: timeout},
		limiter: rate.NewLimiter(rps, burst),
	}
}

// IMPORTANT: create request with context, do not pass context to this function
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	// Wait
	if err := c.limiter.Wait(req.Context()); err != nil {
		return nil, err
	}

	return c.client.Do(req)
}
