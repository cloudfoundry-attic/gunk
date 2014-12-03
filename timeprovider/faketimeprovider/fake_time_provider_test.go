package faketimeprovider_test

import (
	"time"

	"github.com/cloudfoundry/gunk/timeprovider/faketimeprovider"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FakeTimeProvider", func() {
	const Δ time.Duration = 10 * time.Millisecond

	var (
		timeProvider *faketimeprovider.FakeTimeProvider
		initialTime  time.Time
	)

	BeforeEach(func() {
		initialTime = time.Date(2014, 1, 1, 3, 0, 30, 0, time.UTC)
		timeProvider = faketimeprovider.New(initialTime)
	})

	Describe("Time", func() {
		It("returns the current time, w/o race conditions", func() {
			go timeProvider.Increment(time.Minute)
			Eventually(timeProvider.Time).Should(Equal(initialTime.Add(time.Minute)))
		})
	})

	Describe("Sleep", func() {
		It("blocks until the given interval elapses", func() {
			doneSleeping := make(chan struct{})
			go func() {
				timeProvider.Sleep(10 * time.Second)
				close(doneSleeping)
			}()

			Consistently(doneSleeping, Δ).ShouldNot(BeClosed())

			timeProvider.Increment(5 * time.Second)
			Consistently(doneSleeping, Δ).ShouldNot(BeClosed())

			timeProvider.Increment(4 * time.Second)
			Consistently(doneSleeping, Δ).ShouldNot(BeClosed())

			timeProvider.Increment(1 * time.Second)
			Eventually(doneSleeping).Should(BeClosed())
		})
	})
})
