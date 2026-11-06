package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

const readBufferSize = 32 << 20

type HttpClient struct {
	client *http.Client
}

func NewHttpClient(client *http.Client) *HttpClient {
	return &HttpClient{
		client: client,
	}
}

type Response[T any] struct {
	StatusCode int
	Body       T
}

func (h HttpClient) Get[T any](ctx context.Context, url string) (Response[T], error) {
	var out Response[T]

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return out, fmt.Errorf("build request: %w", err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return out, fmt.Errorf("do request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			slog.Error("analytics-service: close response body", "error", cerr)
		}
	}()

	out.StatusCode = resp.StatusCode

	raw, err := io.ReadAll(io.LimitReader(resp.Body, readBufferSize))
	if err != nil {
		return out, fmt.Errorf("read body: %w", err)
	}

	if err := json.Unmarshal(raw, &out.Body); err != nil {
		return out, fmt.Errorf("decode: %w", err)
	}

	return out, nil
}
