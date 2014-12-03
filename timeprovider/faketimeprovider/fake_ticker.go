package faketimeprovider

import (
	"sync"
	"time"

	"github.com/cloudfoundry/gunk/timeprovider"
)

type fakeTicker struct {
	provider *FakeTimeProvider

	mutex    sync.Mutex
	duration time.Duration
	channel  chan time.Time

	timer timeprovider.Timer
}

func NewFakeTicker(provider *FakeTimeProvider, d time.Duration) *fakeTicker {
	channel := make(chan time.Time)
	timer := provider.NewTimer(d)

	go func() {
		for {
			time := <-timer.C()
			timer.Reset(d)
			channel <- time
		}
	}()

	return &fakeTicker{
		provider: provider,
		duration: d,
		channel:  channel,
		timer:    timer,
	}
}

func (ft *fakeTicker) C() <-chan time.Time {
	ft.mutex.Lock()
	defer ft.mutex.Unlock()
	return ft.channel
}

func (ft *fakeTicker) Stop() {
	ft.mutex.Lock()
	ft.timer.Stop()
	ft.mutex.Unlock()
}
