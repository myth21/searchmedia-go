package searchmedia

import (
	"testing"
	"time"
)

// Each call to now() returns the next of the given offsets from a fixed start.
func throttleAt(offsets ...time.Duration) *RequestThrottle {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	throttle := NewRequestThrottle(OneRequestIn)
	throttle.random = func() float64 { return 0.5 }
	throttle.now = func() time.Time {
		offset := offsets[0]
		offsets = offsets[1:]

		return start.Add(offset)
	}

	return throttle
}

func TestThrottleWindow(t *testing.T) {
	throttle := throttleAt(0, time.Second, 5*time.Second)

	if err := throttle.ThrottleByIP("1.1.1.1"); err != nil {
		t.Fatalf("first request: %v", err)
	}
	if err := throttle.ThrottleByIP("1.1.1.1"); err != ErrRequestTooSoon {
		t.Fatalf("second request within the window: got %v", err)
	}
	if err := throttle.ThrottleByIP("1.1.1.1"); err != nil {
		t.Fatalf("request after the window: %v", err)
	}
}

func TestThrottleIPsAreIndependent(t *testing.T) {
	throttle := throttleAt(0, 100*time.Millisecond)

	_ = throttle.ThrottleByIP("1.1.1.1")

	if err := throttle.ThrottleByIP("2.2.2.2"); err != nil {
		t.Fatalf("another ip: %v", err)
	}
}

func TestThrottleGarbageCollection(t *testing.T) {
	for name, roll := range map[string]float64{"evicts on a hit": 0, "keeps on a miss": 0.5} {
		t.Run(name, func(t *testing.T) {
			throttle := throttleAt(0, throttleEntryTTL+time.Second)
			_ = throttle.ThrottleByIP("1.1.1.1")

			throttle.random = func() float64 { return roll }
			_ = throttle.ThrottleByIP("2.2.2.2")

			_, kept := throttle.lastRequestAt["1.1.1.1"]
			if kept != (roll != 0) {
				t.Errorf("1.1.1.1 kept: %v", kept)
			}
			if _, ok := throttle.lastRequestAt["2.2.2.2"]; !ok {
				t.Error("2.2.2.2 was evicted")
			}
		})
	}
}
