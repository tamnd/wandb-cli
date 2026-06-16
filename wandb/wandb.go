// Package wandb is the library behind the wandb command: the HTTP client,
// GraphQL request shaping, and the typed data models for the Weights & Biases
// public GraphQL API.
//
// The public API at api.wandb.ai/graphql is open for reading public workspace
// views (reports), entity profiles, and project metadata -- no API key required.
package wandb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// DefaultUserAgent identifies the client to W&B.
const DefaultUserAgent = "wandb/dev (+https://github.com/tamnd/wandb-cli)"

// Host is the site this client targets.
const Host = "wandb.ai"

// APIEndpoint is the W&B GraphQL endpoint.
const APIEndpoint = "https://api.wandb.ai/graphql"

// ErrNotFound is returned when a query returns null data for a single-record lookup.
var ErrNotFound = errors.New("not found")

// ErrServerError is returned on a GraphQL panic response.
var ErrServerError = errors.New("server error")

// ErrRateLimited is returned after exhausting retries on HTTP 429.
var ErrRateLimited = errors.New("rate limited")

// Config holds constructor parameters for Client.
type Config struct {
	UserAgent string
	Rate      time.Duration
	Retries   int
	Timeout   time.Duration
}

// DefaultConfig returns sensible defaults for the W&B API.
func DefaultConfig() Config {
	return Config{
		UserAgent: DefaultUserAgent,
		Rate:      300 * time.Millisecond,
		Retries:   3,
		Timeout:   30 * time.Second,
	}
}

// Client is a rate-limited GraphQL client for the W&B public API.
type Client struct {
	cfg  Config
	http *http.Client
	mu   sync.Mutex
	last time.Time
}

// NewClient returns a Client configured with cfg.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// gqlRequest is the JSON body for a GraphQL POST.
type gqlRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

// gqlResponse is the envelope returned by the GraphQL API.
type gqlResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []gqlError      `json:"errors,omitempty"`
}

type gqlError struct {
	Message string `json:"message"`
	Path    []any  `json:"path,omitempty"`
}

// postGQL sends a GraphQL query and returns the raw data field.
func (c *Client) postGQL(ctx context.Context, query string, variables map[string]any) (json.RawMessage, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		data, retry, err := c.doGQL(ctx, query, variables)
		if err == nil {
			return data, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("graphql: %w", lastErr)
}

func (c *Client) doGQL(ctx context.Context, query string, variables map[string]any) (json.RawMessage, bool, error) {
	c.pace()

	body, err := json.Marshal(gqlRequest{Query: query, Variables: variables})
	if err != nil {
		return nil, false, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, APIEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, true, ErrRateLimited
	}
	if resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, true, err
	}

	var gr gqlResponse
	if err := json.Unmarshal(b, &gr); err != nil {
		return nil, false, fmt.Errorf("decode graphql response: %w", err)
	}

	// GraphQL errors: server panics or schema mismatches.
	if len(gr.Errors) > 0 {
		msg := gr.Errors[0].Message
		if strings.Contains(msg, "panic occurred") {
			return nil, false, ErrServerError
		}
		return nil, false, fmt.Errorf("graphql: %s", msg)
	}

	return gr.Data, false, nil
}

// pace blocks until at least Rate has passed since the previous request.
func (c *Client) pace() {
	if c.cfg.Rate <= 0 {
		return
	}
	c.mu.Lock()
	wait := c.cfg.Rate - time.Since(c.last)
	c.mu.Unlock()
	if wait > 0 {
		time.Sleep(wait)
	}
	c.mu.Lock()
	c.last = time.Now()
	c.mu.Unlock()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}
