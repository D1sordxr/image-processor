package options

import (
	"time"

	"github.com/wb-go/wbf/retry"
)

type RetryStrategy = retry.Strategy

var (
	DefaultStrategy = RetryStrategy{Attempts: 3, Delay: time.Second}
	BrokerStrategy  = RetryStrategy{Attempts: 5, Delay: 3 * time.Second}
)
