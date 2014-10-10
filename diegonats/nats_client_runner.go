package diegonats

import (
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pivotal-golang/lager"
)

type NATSClientRunner struct {
	addresses string
	username  string
	password  string
	logger    lager.Logger
	client    NATSClient
}

func NewClientRunner(addresses, username, password string, logger lager.Logger, client NATSClient) NATSClientRunner {
	return NATSClientRunner{
		addresses: addresses,
		username:  username,
		password:  password,
		logger:    logger.Session("nats-runner"),
		client:    client,
	}
}

func (runner NATSClientRunner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	natsMembers := []string{}
	for _, addr := range strings.Split(runner.addresses, ",") {
		uri := url.URL{
			Scheme: "nats",
			User:   url.UserPassword(runner.username, runner.password),
			Host:   addr,
		}
		natsMembers = append(natsMembers, uri.String())
	}

	err := runner.client.Connect(natsMembers)
	for err != nil {
		runner.logger.Error("connecting-to-nats-failed", err)
		select {
		case <-signals:
			return nil
		case <-time.After(time.Second):
			err = runner.client.Connect(natsMembers)
		}
	}

	runner.logger.Info("connecting-to-nats-succeeeded")
	close(ready)

	<-signals
	runner.client.Disconnect()
	return nil
}
