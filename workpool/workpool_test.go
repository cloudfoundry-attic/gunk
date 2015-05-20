package workpool_test

import (
	"sync"
	"sync/atomic"
	"time"

	. "github.com/cloudfoundry/gunk/workpool"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Workpool", func() {
	var pool *WorkPool
	var poolSize int
	var pendingSize int

	BeforeEach(func() {
		poolSize = 2
	})

	AfterEach(func() {
		if pool != nil {
			pool.Stop()
		}
	})

	Context("when the number of workers is non-positive", func() {
		BeforeEach(func() {
			poolSize = 0
		})

		It("errors", func() {
			_, err := New(poolSize, pendingSize)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when the number of allowed pending workers is negative", func() {
		BeforeEach(func() {
			pendingSize = -1
		})

		It("errors", func() {
			_, err := New(poolSize, pendingSize)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("scheduling work", func() {
		Context("no pending work allowed", func() {
			BeforeEach(func() {
				pendingSize = 0
			})

			Context("when passed one work", func() {
				It("should run the passed in function", func() {
					var err error
					pool, err = New(poolSize, pendingSize)
					Expect(err).NotTo(HaveOccurred())

					called := make(chan bool)

					pool.Submit(func() {
						called <- true
					})

					Eventually(called, 0.1, 0.01).Should(Receive())
				})
			})

			Context("when passed many work", func() {
				var (
					startTime time.Time
					runTimes  chan time.Duration
					sleepTime time.Duration
					work      func()
				)

				BeforeEach(func() {
					startTime = time.Now()
					runTimes = make(chan time.Duration, 10)
					sleepTime = time.Duration(0.01 * float64(time.Second))

					work = func() {
						time.Sleep(sleepTime)
						runTimes <- time.Since(startTime)
					}
				})

				Context("when passed poolSize work", func() {
					It("should run the functions concurrently", func() {
						var err error
						pool, err = New(poolSize, pendingSize)
						Expect(err).NotTo(HaveOccurred())

						pool.Submit(work)
						pool.Submit(work)

						Eventually(runTimes, 0.1, 0.01).Should(HaveLen(2))
						Expect(<-runTimes).To(BeNumerically("<=", sleepTime+sleepTime/2))
						Expect(<-runTimes).To(BeNumerically("<=", sleepTime+sleepTime/2))
					})
				})

				Context("when passed more than poolSize work", func() {
					It("should run all the functions, but at most poolSize at a time", func() {
						var err error
						pool, err = New(poolSize, pendingSize)
						Expect(err).NotTo(HaveOccurred())

						pool.Submit(work)
						pool.Submit(work)
						pool.Submit(work)

						Eventually(runTimes, 0.1, 0.01).Should(HaveLen(3))

						//first batch
						Expect(<-runTimes).To(BeNumerically("<=", sleepTime+sleepTime/2))
						Expect(<-runTimes).To(BeNumerically("<=", sleepTime+sleepTime/2))

						//second batch
						Expect(<-runTimes).To(BeNumerically(">=", sleepTime*2))
					})
				})
			})
		})

		Context("pending work allowed", func() {
			BeforeEach(func() {
				pendingSize = 1
			})

			Context("when passed more than poolSize work", func() {
				It("should not block the caller", func() {
					var err error
					pool, err = New(poolSize, pendingSize)
					Expect(err).NotTo(HaveOccurred())

					barrier := make(chan struct{})
					wg := sync.WaitGroup{}

					work := func() {
						wg.Done()
						<-barrier
					}

					defer close(barrier)

					wg.Add(2)
					pool.Submit(work)
					pool.Submit(work)

					wg.Wait()

					var count int32
					go func() {
						pool.Submit(func() {
							defer GinkgoRecover()
							Expect(atomic.CompareAndSwapInt32(&count, 1, 2)).To(BeTrue())
						})
						Expect(atomic.CompareAndSwapInt32(&count, 0, 1)).To(BeTrue())
					}()

					Eventually(func() int32 { return atomic.LoadInt32(&count) }).Should(Equal(int32(1)))
					barrier <- struct{}{}

					Eventually(func() int32 { return atomic.LoadInt32(&count) }).Should(Equal(int32(2)))
				})
			})
		})

		Context("when stopped", func() {
			It("should never perform the work", func() {
				var err error
				pool, err = New(poolSize, pendingSize)
				Expect(err).NotTo(HaveOccurred())

				pool.Stop()

				called := make(chan bool, 1)
				pool.Submit(func() {
					called <- true
				})

				Consistently(called).ShouldNot(Receive())
			})

			It("should stop the workers", func() {
				var err error
				pool, err = New(poolSize, pendingSize)
				Expect(err).NotTo(HaveOccurred())

				called := make(chan bool)
				pool.Submit(func() {
					called <- true
				})

				Eventually(called).Should(Receive())

				pool.Stop()
				Eventually(func() int {
					_, active := pool.Stats()
					return active
				}).Should(Equal(0), "Should have reduced number of go routines by pool size")
			})
		})
	})
})
