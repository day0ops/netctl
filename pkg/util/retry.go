package util

import (
	"time"

	"github.com/cenkalti/backoff/v4"

	"github.com/day0ops/netctl/pkg/log"
)

// LocalRetry is back-off retry for local connections
func LocalRetry(callback func() error, maxTime time.Duration) error {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 250 * time.Millisecond
	b.RandomizationFactor = 0.25
	b.Multiplier = 1.25
	b.MaxElapsedTime = maxTime
	return backoff.RetryNotify(callback, b, notify)
}

func notify(err error, d time.Duration) {
	log.Infof("will retry after %s: %v", d, err)
}
