package timeprovider

import "time"

type TimeProvider interface {
	Now() time.Time
	Sleep(d time.Duration)

	NewTimer(d time.Duration) Timer
	NewTicker(d time.Duration) Ticker
}

type RealTimeProvider struct{}

func NewTimeProvider() (provider *RealTimeProvider) {
	return &RealTimeProvider{}
}

func (provider *RealTimeProvider) Now() time.Time {
	return time.Now()
}

func (provider *RealTimeProvider) Sleep(d time.Duration) {
	<-provider.NewTimer(d).C()
}

func (provider *RealTimeProvider) NewTimer(d time.Duration) Timer {
	return &realTimer{
		t: time.NewTimer(d),
	}
}

func (provider *RealTimeProvider) NewTicker(d time.Duration) Ticker {
	return &realTicker{
		t: time.NewTicker(d),
	}
}
