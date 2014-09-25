package natsclientrunner_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/cloudfoundry/gunk/natsclientrunner"
	"github.com/cloudfoundry/yagnats"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"

	"github.com/cloudfoundry/gunk/natsrunner"
	"github.com/tedsuo/ifrit"
)

var natsRunner *natsrunner.NATSRunner
var natsPort int

func TestListener(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NatsClientRunner Suite")
}

var _ = BeforeSuite(func() {
	natsPort = 4001 + GinkgoParallelNode()
	natsRunner = natsrunner.NewNATSRunner(natsPort)
	natsRunner.Start()
})

var _ = AfterSuite(func() {
	natsRunner.Stop()
})

var _ = Describe("Starting the NatsClientRunner process", func() {
	var natsClient yagnats.NATSConn
	var natsClientRunner ifrit.Runner
	var natsClientProcess ifrit.Process

	BeforeEach(func() {
		natsAddress := fmt.Sprintf("127.0.0.1:%d", natsPort)
		natsClientRunner = New(natsAddress, "nats", "nats", lagertest.NewTestLogger("test"), &natsClient)
	})

	AfterEach(func() {
		if natsClientProcess != nil {
			natsClientProcess.Signal(os.Interrupt)
			Eventually(natsClientProcess.Wait(), 5).Should(Receive())
		}
	})

	Describe("when NATS is up", func() {
		BeforeEach(func() {
			natsClientProcess = ifrit.Envoke(natsClientRunner)
		})

		It("connects to NATS", func() {
			Ω(natsClient.Ping()).Should(BeTrue())
		})

		It("disconnects when it receives a signal", func() {
			natsClientProcess.Signal(os.Interrupt)
			Eventually(natsClientProcess.Wait(), 5).Should(Receive())

			Ω(natsClient.Ping()).Should(BeFalse())
		})
	})

	Describe("when NATS is not up", func() {
		var natsClientProcessChan chan ifrit.Process

		BeforeEach(func() {
			natsRunner.Stop()

			natsClientProcessChan = make(chan ifrit.Process, 1)
			go func() {
				natsClientProcessChan <- ifrit.Envoke(natsClientRunner)
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
