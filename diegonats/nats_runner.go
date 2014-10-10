package diegonats

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var natsCommand *exec.Cmd

type NATSRunner struct {
	Client NATSClient

	port        int
	natsSession *gexec.Session
	natsUrls    []string
}

func NewRunner(port int) *NATSRunner {
	return &NATSRunner{
		port: port,
	}
}

func (runner *NATSRunner) Start() {
	if runner.natsSession != nil {
		panic("starting an already started NATS runner!!!")
	}

	_, err := exec.LookPath("gnatsd")
	if err != nil {
		fmt.Println("You need gnatsd installed!")
		os.Exit(1)
	}

	cmd := exec.Command("gnatsd", "-p", strconv.Itoa(runner.port))
	sess, err := gexec.Start(
		cmd,
		gexec.NewPrefixedWriter("\x1b[32m[o]\x1b[34m[gnatsd]\x1b[0m ", ginkgo.GinkgoWriter),
		gexec.NewPrefixedWriter("\x1b[91m[e]\x1b[34m[gnatsd]\x1b[0m ", ginkgo.GinkgoWriter),
	)
	Ω(err).ShouldNot(HaveOccurred(), "Make sure to have gnatsd on your path")

	runner.natsSession = sess

	Expect(err).NotTo(HaveOccurred())

	natsClient := NewClient()
	Eventually(func() error {
		err = natsClient.Connect([]string{fmt.Sprintf("nats://127.0.0.1:%d", runner.port)})
		return err
	}, 5, 0.1).ShouldNot(HaveOccurred())

	runner.Client = natsClient
}

func (runner *NATSRunner) Stop() {
	runner.KillWithFire()
}

func (runner *NATSRunner) KillWithFire() {
	runner.Client = nil

	if runner.natsSession != nil {
		runner.natsSession.Kill().Wait(5 * time.Second)
		runner.natsSession = nil
	}
}
