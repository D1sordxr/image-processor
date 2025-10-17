package ticker

import (
	"time"

	"github.com/D1sordxr/image-processor/internal/domain/core/image/options"
)

func NewHealth() *time.Ticker {
	return time.NewTicker(options.HealthTickerInterval)
}
