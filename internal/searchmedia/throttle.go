package searchmedia

import (
	"errors"
	"math/rand/v2"
	"sync"
	"time"
)

const (
	OneRequestIn            = 5 * time.Second
	throttleEntryTTL        = time.Hour
	garbageCollectionChance = 1.0 / 100
)

var ErrRequestTooSoon = errors.New("more than one request in the time window by ip")

type RequestThrottle struct {
	window        time.Duration
	mutex         sync.Mutex
	lastRequestAt map[string]time.Time

	// These functions can be replaced in tests to control time and randomness.
	now    func() time.Time
	random func() float64
}

func NewRequestThrottle(window time.Duration) *RequestThrottle {
	return &RequestThrottle{
		window:        window,
		lastRequestAt: map[string]time.Time{},
		now:           time.Now,
		random:        rand.Float64,
	}
}

func (t *RequestThrottle) ThrottleByIP(ip string) error {
	now := t.now()

	t.mutex.Lock()
	defer t.mutex.Unlock()

	if last, seen := t.lastRequestAt[ip]; seen && now.Sub(last) < t.window {
		return ErrRequestTooSoon
	}

	t.lastRequestAt[ip] = now

	if t.random() < garbageCollectionChance {
		t.cleanupExpiredEntries(now)
	}

	return nil
}

func (t *RequestThrottle) cleanupExpiredEntries(now time.Time) {
	for ip, at := range t.lastRequestAt {
		if now.Sub(at) > throttleEntryTTL {
			delete(t.lastRequestAt, ip)
		}
	}
}