package faketimeprovider_test

import (
	"time"

	"github.com/cloudfoundry/gunk/timeprovider/faketimeprovider"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FakeTimer", func() {
	const Δ = 10 * time.Millisecond

	var (
		timeProvider *faketimeprovider.FakeTimeProvider
		initialTime  time.Time
	)

	BeforeEach(func() {
		initialTime = time.Date(2014, 1, 1, 3, 0, 30, 0, time.UTC)
		timeProvider = faketimeprovider.New(initialTime)
	})

	It("proivdes a channel that receives after the given interval has elapsed", func() {
		timer := timeProvider.NewTimer(10 * time.Second)
		timeChan := timer.C()
		Consistently(timeChan, Δ).ShouldNot(Receive())

		timeProvider.Increment(5 * time.Second)
		Consistently(timeChan, Δ).ShouldNot(Receive())

		timeProvider.Increment(4 * time.Second)
		Consistently(timeChan, Δ).ShouldNot(Receive())

		timeProvider.Increment(1 * time.Second)
		Eventually(timeChan).Should(Receive(Equal(initialTime.Add(10 * time.Second))))

		timeProvider.Increment(10 * time.Second)
		Consistently(timeChan, Δ).ShouldNot(Receive())
	})
})
