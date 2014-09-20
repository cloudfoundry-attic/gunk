package natsclientrunner_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/cloudfoundry/gunk/natsclientrunner"
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
	var natsClientRunner ifrit.Runner

	BeforeEach(func() {
		natsAddress := fmt.Sprintf("127.0.0.1:%d", natsPort)
		natsClient := NewClient(natsAddress, "nats", "nats")
		natsClientRunner = New(natsClient, lagertest.NewTestLogger("test"))
	})

	Describe("when NATS is up", func() {
		It("waits for NATS to come up", func() {
			natsClientProcess := ifrit.Envoke(natsClientRunner)
			natsClientProcess.Signal(os.Interrupt)

			// wait for process to terminate
			Eventually(natsClientProcess.Wait(), 5).Should(Receive())
		})
	})

	Describe("when NATS is not up", func() {
		Context("and the stager is started", func() {
			var natsClientProcessChan chan ifrit.Process

			BeforeEach(func() {
				natsRunner.Stop()

				natsClientProcessChan = make(chan ifrit.Process, 1)
				go func() {
					natsClientProcessChan <- ifrit.Envoke(natsClientRunner)
				}()
			})

			It("waits for NATS to come up", func() {
				Consistently(natsClientProcessChan).ShouldNot(Receive())

				natsRunner.Start()

				var natsClientProcess ifrit.Process
				Eventually(natsClientProcessChan, 5*time.Second).Should(Receive(&natsClientProcess))

				natsClientProcess.Signal(os.Interrupt)
				Eventually(natsClientProcess.Wait(), 5).Should(Receive())
			})
		})
	})
})
