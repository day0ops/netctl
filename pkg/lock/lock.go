package lock

import (
	"crypto/sha1"
	"fmt"
	"time"

	"github.com/juju/clock"
	"github.com/juju/mutex/v2"
)

// PathMutexSpec returns a mutex spec for a path
func PathMutexSpec(path string) mutex.Spec {
	s := mutex.Spec{
		Name:  fmt.Sprintf("mk%x", sha1.Sum([]byte(path)))[0:40],
		Clock: clock.WallClock,
		// Poll the lock twice a second
		Delay: 500 * time.Millisecond,
		// panic after a minute instead of locking infinitely
		Timeout: 60 * time.Second,
	}
	return s
}
