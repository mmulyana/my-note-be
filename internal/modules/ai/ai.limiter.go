package ai

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	limitRequests = 15
	limitWindow   = time.Minute
)

var (
	ErrRateLimited = errors.New("too many ai requests, try again in a moment")
	ErrStreamBusy  = errors.New("another ai request is still running")
)

type limiter struct {
	mu        sync.Mutex
	hits      map[uuid.UUID][]time.Time
	active    map[uuid.UUID]struct{}
	lastSweep time.Time
	max       int
	window    time.Duration
}

func newLimiter(max int, window time.Duration) *limiter {
	return &limiter{
		hits:   make(map[uuid.UUID][]time.Time),
		active: make(map[uuid.UUID]struct{}),
		max:    max,
		window: window,
	}
}

func (l *limiter) acquire(id uuid.UUID) (release func(), retryAfter time.Duration, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	l.sweep(now)

	if _, busy := l.active[id]; busy {
		return nil, time.Second, ErrStreamBusy
	}

	recent := l.recent(id, now)
	if len(recent) >= l.max {
		l.hits[id] = recent
		return nil, recent[0].Add(l.window).Sub(now), ErrRateLimited
	}

	l.hits[id] = append(recent, now)
	l.active[id] = struct{}{}

	return func() {
		l.mu.Lock()
		defer l.mu.Unlock()
		delete(l.active, id)
	}, 0, nil
}

func (l *limiter) recent(id uuid.UUID, now time.Time) []time.Time {
	cutoff := now.Add(-l.window)
	hits := l.hits[id]
	i := 0
	for i < len(hits) && !hits[i].After(cutoff) {
		i++
	}
	return hits[i:]
}

func (l *limiter) sweep(now time.Time) {
	if now.Sub(l.lastSweep) < l.window {
		return
	}
	l.lastSweep = now
	for id := range l.hits {
		if recent := l.recent(id, now); len(recent) == 0 {
			delete(l.hits, id)
		} else {
			l.hits[id] = recent
		}
	}
}
