package runtimestate

import "sync"

var (
	mu      sync.RWMutex
	started bool
)

func SetStarted(value bool) {
	mu.Lock()
	defer mu.Unlock()
	started = value
}

func Started() bool {
	mu.RLock()
	defer mu.RUnlock()
	return started
}
