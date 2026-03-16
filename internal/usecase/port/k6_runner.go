package port

import (
	"context"
	"io"

	"github.com/hironow/dominator/internal/domain"
)

// K6Runner executes k6 load test scripts and returns parsed results.
type K6Runner interface {
	Run(ctx context.Context, scriptPath string, load domain.LoadConfig, stderrW io.Writer) (domain.K6Results, error)
	Validate(ctx context.Context, scriptPath string) error
}
