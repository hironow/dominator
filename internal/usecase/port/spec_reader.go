package port

import "context"

// SpecReader fetches API specification content from a URL.
type SpecReader interface {
	Fetch(ctx context.Context, url string) ([]byte, error)
}
