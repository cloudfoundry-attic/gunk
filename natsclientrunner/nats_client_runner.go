package natsclientrunner

import (
	"os"
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
