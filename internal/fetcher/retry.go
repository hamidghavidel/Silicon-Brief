package fetcher

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// doWithRetry performs an HTTP request with exponential backoff retries.
func doWithRetry(ctx context.Context, req *http.Request, maxRetries int) (*http.Response, error) {
	var resp *http.Response
	var err error

	for i := 0; i <= maxRetries; i++ {
		resp, err = http.DefaultClient.Do(req)
		if err == nil && resp.StatusCode < 500 {
			return resp, nil
		}
		if err == nil {
			resp.Body.Close()
		}

		if i < maxRetries {
			delay := time.Duration(1<<i) * time.Second
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", maxRetries, err)
	}
	return nil, fmt.Errorf("server error after %d retries: status %d", maxRetries, resp.StatusCode)
}
