package httpclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/forgedash/forgedash/internal/domain/shared"
)

type Client struct {
	httpClient *http.Client
	maxRetries int
}

func NewClient(timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		maxRetries: 2,
	}
}

func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)

	var resp *http.Response
	var err error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		resp, err = c.httpClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}

			if attempt < c.maxRetries {
				time.Sleep(time.Duration(200*(1<<attempt)) * time.Millisecond)
				continue
			}
			return nil, fmt.Errorf("%w: %v", shared.ErrProviderUnavailable, err)
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfterHeader := resp.Header.Get("Retry-After")
			waitSeconds := 1
			if retryAfterHeader != "" {
				if sec, parseErr := strconv.Atoi(retryAfterHeader); parseErr == nil && sec > 0 {
					waitSeconds = sec
				}
			}
			resp.Body.Close()
			if attempt < c.maxRetries && waitSeconds <= 3 {
				select {
				case <-time.After(time.Duration(waitSeconds) * time.Second):
					continue
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
			return nil, shared.ErrRateLimited
		}

		if resp.StatusCode >= 500 && resp.StatusCode <= 504 {
			resp.Body.Close()
			if attempt < c.maxRetries {
				select {
				case <-time.After(time.Duration(250*(1<<attempt)) * time.Millisecond):
					continue
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
			return nil, shared.ErrProviderUnavailable
		}

		switch resp.StatusCode {
		case http.StatusUnauthorized:
			resp.Body.Close()
			return nil, shared.ErrUnauthorized
		case http.StatusForbidden:
			resp.Body.Close()
			return nil, shared.ErrForbidden
		case http.StatusNotFound:
			resp.Body.Close()
			return nil, shared.ErrNotFound
		}

		return resp, nil
	}

	return nil, err
}

func DrainAndClose(r io.ReadCloser) {
	if r != nil {
		_, _ = io.Copy(io.Discard, r)
		_ = r.Close()
	}
}
