package faketimeprovider_test

import (
	"time"

	"github.com/cloudfoundry/gunk/timeprovider/faketimeprovider"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FakeTicker", func() {
	const Δ = 10 * time.Millisecond

	var (
		timeProvider *faketimeprovider.FakeTimeProvider
		initialTime  time.Time
	)

	BeforeEach(func() {
		initialTime = time.Date(2014, 1, 1, 3, 0, 30, 0, time.UTC)
		timeProvider = faketimeprovider.New(initialTime)
	})

	It("provides a channel that receives the time at each interval", func() {
		ticker := timeProvider.NewTicker(10 * time.Second)
		timeChan := ticker.C()
		Consistently(timeChan, Δ).ShouldNot(Receive())

		timeProvider.Increment(5 * time.Second)
		Consistently(timeChan, Δ).ShouldNot(Receive())

		timeProvider.Increment(4 * time.Second)
		Consistently(timeChan, Δ).ShouldNot(Receive())

		timeProvider.Increment(1 * time.Second)
		Eventually(timeChan).Should(Receive(Equal(initialTime.Add(10 * time.Second))))

		timeProvider.Increment(10 * time.Second)
		Eventually(timeChan).Should(Receive(Equal(initialTime.Add(20 * time.Second))))

		timeProvider.Increment(10 * time.Second)
		Eventually(timeChan).Should(Receive(Equal(initialTime.Add(30 * time.Second))))
	})
})
