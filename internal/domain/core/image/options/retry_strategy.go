package options

import (
	"github.com/wb-go/wbf/retry"
	"time"
)

type RetryStrategy = retry.Strategy

var (
	DefaultStrategy = RetryStrategy{Attempts: 3, Delay: time.Second}
	BrokerStrategy  = RetryStrategy{Attempts: 5, Delay: 3 * time.Second}
)
