package workpool_test

import (
	"github.com/cloudfoundry/gunk/workpool"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Throttler", func() {
	var throttler *workpool.Throttler

	Context("when max workers is non-positive", func() {
		It("errors", func() {
			_, err := workpool.NewThrottler(0, []func(){})
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when max workers is positive", func() {
		var maxWorkers int
		var calledChan, unblockChan chan struct{}
		var work func()

		BeforeEach(func() {
			maxWorkers = 2
			calledChan = make(chan struct{})
			unblockChan = make(chan struct{})
			work = func() {
				calledChan := calledChan
				unblockChan := unblockChan
				calledChan <- struct{}{}
				<-unblockChan
			}
		})

		AfterEach(func() {
			close(calledChan)
			close(unblockChan)
		})

		Describe("Work", func() {
			Context("when requesting less work than the max number of workers", func() {
				BeforeEach(func() {
					works := make([]func(), maxWorkers-1)
					for i := range works {
						works[i] = work
					}

					var err error
					throttler, err = workpool.NewThrottler(maxWorkers, works)
					Expect(err).NotTo(HaveOccurred())
				})

				It("should run the passed-in work", func() {
					go throttler.Work()

					for i := 0; i < maxWorkers-1; i++ {
						Eventually(calledChan).Should(Receive())
					}
				})
			})

			Context("when submitting work equal to the number of workers", func() {
				BeforeEach(func() {
					works := make([]func(), maxWorkers)
					for i := range works {
						works[i] = work
					}

					var err error
					throttler, err = workpool.NewThrottler(maxWorkers, works)
					Expect(err).NotTo(HaveOccurred())
				})

				It("should run the passed-in work concurrently", func() {
					go throttler.Work()

					for i := 0; i < maxWorkers; i++ {
						Eventually(calledChan).Should(Receive())
					}
				})
			})

			Context("when submitting more work than the max number of workers", func() {
				BeforeEach(func() {
					works := make([]func(), maxWorkers+1)
					for i := range works {
						works[i] = work
					}

					var err error
					throttler, err = workpool.NewThrottler(maxWorkers, works)
					Expect(err).NotTo(HaveOccurred())
				})

				It("should run the passed-in work concurrently up to the max number of workers at a time", func() {
					go throttler.Work()

					for i := 0; i < maxWorkers; i++ {
						Eventually(calledChan).Should(Receive())
					}
					Consistently(calledChan).ShouldNot(Receive())

					unblockChan <- struct{}{}

					Eventually(calledChan).Should(Receive())
				})
			})
		})
	})
})
