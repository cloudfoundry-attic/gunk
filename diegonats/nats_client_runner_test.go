package diegonats_test

import (
	"errors"
	"fmt"
	"os"
	"time"

	. "github.com/cloudfoundry/gunk/diegonats"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"

	"github.com/tedsuo/ifrit"
)

var _ = Describe("Starting the NatsClientRunner process", func() {
	var natsClient NATSClient
	var natsClientRunner ifrit.Runner
	var natsClientProcess ifrit.Process

	BeforeEach(func() {
		natsAddress := fmt.Sprintf("127.0.0.1:%d", natsPort)
		natsClient = NewClient()
		natsClientRunner = NewClientRunner(natsAddress, "nats", "nats", lagertest.NewTestLogger("test"), natsClient)
	})

	AfterEach(func() {
		if natsClientProcess != nil {
			natsClientProcess.Signal(os.Interrupt)
			Eventually(natsClientProcess.Wait(), 5).Should(Receive())
		}
	})

	Describe("when NATS is up", func() {
		BeforeEach(func() {
			natsClientProcess = ifrit.Invoke(natsClientRunner)
		})

		It("connects to NATS", func() {
			Ω(natsClient.Ping()).Should(BeTrue())
		})

		It("disconnects when it receives a signal", func() {
			natsClientProcess.Signal(os.Interrupt)
			Eventually(natsClientProcess.Wait(), 5).Should(Receive())

			Ω(natsClient.Ping()).Should(BeFalse())
		})

		It("exits with an error when nats connection is closed permanently", func() {
			errorChan := natsClientProcess.Wait()

			natsClient.Disconnect()

			Eventually(errorChan).Should(Receive(Equal(errors.New("nats closed unexpectedly"))))
		})
	})

	Describe("when NATS is not up", func() {
		var natsClientProcessChan chan ifrit.Process

		BeforeEach(func() {
			natsRunner.Stop()

			natsClientProcessChan = make(chan ifrit.Process, 1)
			go func() {
				natsClientProcessChan <- ifrit.Invoke(natsClientRunner)
			}()
		})

		It("waits for NATS to come up and connects to NATS", func() {
			Consistently(natsClientProcessChan).ShouldNot(Receive())
			natsRunner.Start()
			Eventually(natsClientProcessChan, 5*time.Second).Should(Receive(&natsClientProcess))

			Ω(natsClient.Ping()).Should(BeTrue())
		})

		It("disconnects when it receives a signal", func() {
			Consistently(natsClientProcessChan).ShouldNot(Receive())
			natsRunner.Start()
			Eventually(natsClientProcessChan, 5*time.Second).Should(Receive(&natsClientProcess))

			natsClientProcess.Signal(os.Interrupt)
			Eventually(natsClientProcess.Wait(), 5).Should(Receive())

			Ω(natsClient.Ping()).Should(BeFalse())
		})
	})
})
