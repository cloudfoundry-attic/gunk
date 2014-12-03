package faketimeprovider

import (
	"sync"
	"time"

	"github.com/cloudfoundry/gunk/timeprovider"
)

type timeWatcher interface {
	timeUpdated(time.Time)
}

type FakeTimeProvider struct {
	sync.Mutex
	now time.Time

	watchers map[timeWatcher]struct{}
}

func New(now time.Time) *FakeTimeProvider {
	return &FakeTimeProvider{
		now:      now,
		watchers: make(map[timeWatcher]struct{}),
	}
}

func (provider *FakeTimeProvider) Time() time.Time {
	return provider.Now()
}

func (provider *FakeTimeProvider) Now() time.Time {
	provider.Mutex.Lock()
	defer provider.Mutex.Unlock()

	return provider.now
}

func (provider *FakeTimeProvider) Increment(duration time.Duration) {
	provider.Mutex.Lock()
	now := provider.now.Add(duration)
	provider.now = now

	watchers := make([]timeWatcher, 0, len(provider.watchers))
	for w, _ := range provider.watchers {
		watchers = append(watchers, w)
	}
	provider.Mutex.Unlock()

	for _, w := range watchers {
		w.timeUpdated(now)
	}
}

func (provider *FakeTimeProvider) IncrementBySeconds(seconds uint64) {
	provider.Increment(time.Duration(seconds) * time.Second)
}

func (provider *FakeTimeProvider) NewTimer(d time.Duration) timeprovider.Timer {
	timer := NewFakeTimer(provider, d)
	provider.addTimeWatcher(timer)

	return timer
}

func (provider *FakeTimeProvider) Sleep(d time.Duration) {
	<-provider.NewTimer(d).C()
}

func (provider *FakeTimeProvider) NewTicker(d time.Duration) timeprovider.Ticker {
	return NewFakeTicker(provider, d)
}

func (provider *FakeTimeProvider) WatcherCount() int {
	provider.Mutex.Lock()
	defer provider.Mutex.Unlock()

	return len(provider.watchers)
}

func (provider *FakeTimeProvider) addTimeWatcher(tw timeWatcher) {
	provider.Mutex.Lock()
	provider.watchers[tw] = struct{}{}
	provider.Mutex.Unlock()

	tw.timeUpdated(provider.Time())
}

func (provider *FakeTimeProvider) removeTimeWatcher(tw timeWatcher) {
	provider.Mutex.Lock()
	delete(provider.watchers, tw)
	provider.Mutex.Unlock()
}
