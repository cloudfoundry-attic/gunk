package diegonats_test

import (
	"testing"

	"github.com/cloudfoundry/gunk/diegonats"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDiegoNATS(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Diego NATS Suite")
}

var natsRunner *diegonats.NATSRunner
var natsPort int

var _ = BeforeSuite(func() {
	natsPort = 4001 + GinkgoParallelNode()
	natsRunner = diegonats.NewRunner(natsPort)
	natsRunner.Start()
})

var _ = AfterSuite(func() {
	natsRunner.Stop()
})
