package natsclientrunner

import (
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/cloudfoundry/yagnats"
	"github.com/pivotal-golang/lager"
)

type Runner struct {
	client yagnats.ApceraWrapperNATSClient
	logger lager.Logger
}

func New(client yagnats.ApceraWrapperNATSClient, logger lager.Logger) Runner {
	return Runner{
		client: client,
		logger: logger.Session("nats-runner"),
	}
}

func NewClient(addresses, username, password string) yagnats.ApceraWrapperNATSClient {
	natsMembers := []string{}
	for _, addr := range strings.Split(addresses, ",") {
		uri := url.URL{
			Scheme: "nats",
			User:   url.UserPassword(username, password),
			Host:   addr,
		}
		natsMembers = append(natsMembers, uri.String())
	}

	return yagnats.NewApceraClientWrapper(natsMembers)
}

func (c Runner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	err := c.client.Connect()
	for err != nil {
		c.logger.Error("connecting-to-nats-failed", err)
		select {
		case <-signals:
			return nil
		case <-time.After(time.Second):
			err = c.client.Connect()
		}
	}

	c.logger.Info("connecting-to-nats-succeeeded")
	close(ready)

	<-signals
	return nil
}
