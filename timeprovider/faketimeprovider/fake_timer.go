package faketimeprovider

import (
	"sync"
	"time"
)

type fakeTimer struct {
	provider *FakeTimeProvider

	mutex          sync.Mutex
	completionTime time.Time
	channel        chan time.Time
}

func NewFakeTimer(provider *FakeTimeProvider, d time.Duration) *fakeTimer {
	return &fakeTimer{
		provider:       provider,
		completionTime: provider.Now().Add(d),
		channel:        make(chan time.Time, 1),
	}
}

func (ft *fakeTimer) C() <-chan time.Time {
	ft.mutex.Lock()
	defer ft.mutex.Unlock()
	return ft.channel
}

func (ft *fakeTimer) Reset(d time.Duration) bool {
	currentTime := ft.provider.Now()

	ft.mutex.Lock()
	active := !ft.completionTime.IsZero()
	ft.completionTime = currentTime.Add(d)
	ft.mutex.Unlock()

	ft.provider.addTimeWatcher(ft)

	return active
}

func (ft *fakeTimer) Stop() bool {
	ft.mutex.Lock()
	active := !ft.completionTime.IsZero()
	ft.completionTime = time.Time{}
	ft.mutex.Unlock()

	ft.provider.removeTimeWatcher(ft)

	return active
}

func (ft *fakeTimer) timeUpdated(now time.Time) {
	var fire bool

	ft.mutex.Lock()
	if !ft.completionTime.IsZero() {
		fire = now.After(ft.completionTime) || now.Equal(ft.completionTime)
	}
	ft.mutex.Unlock()

	if fire {
		select {
		case ft.channel <- now:
			ft.Stop()
		default:
		}
	}
}
