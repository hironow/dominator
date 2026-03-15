package session

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// HTTPSpecReader implements port.SpecReader via HTTP GET.
type HTTPSpecReader struct{}

// Fetch retrieves the API specification content from the given URL.
func (r *HTTPSpecReader) Fetch(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch spec: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("spec fetch failed: HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read spec body: %w", err)
	}
	return data, nil
}
