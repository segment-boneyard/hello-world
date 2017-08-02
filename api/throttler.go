package api

import "time"

// Throttler can be used to apply client-side throttling
type Throttler struct {
	eventsPerDuration int
	duration          time.Duration
	interval          time.Duration
	queue             chan struct{}
}

func (t *Throttler) run() {
	interval := time.Duration(float64(t.duration) / float64(t.eventsPerDuration))
	for {
		select {
		case t.queue <- struct{}{}:
		default:
		}
		time.Sleep(interval)
	}
}

// Use will block until the next event is allowed to start
func (t *Throttler) Use() {
	<-t.queue
}

// NewThrottler(50, time.Second) will return a throttler allowing no more than 50 events start
// during a given second
func NewThrottler(eventsPerDuration int, duration time.Duration) *Throttler {
	t := Throttler{
		eventsPerDuration: eventsPerDuration,
		duration:          duration,
		queue:             make(chan struct{}, eventsPerDuration),
	}
	go t.run()
	return &t
}
