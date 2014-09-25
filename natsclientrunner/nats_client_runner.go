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
	addresses string
	username  string
	password  string
	logger    lager.Logger
	client    *yagnats.NATSConn
}

func New(addresses, username, password string, logger lager.Logger, client *yagnats.NATSConn) Runner {
	return Runner{
		addresses: addresses,
		username:  username,
		password:  password,
		logger:    logger.Session("nats-runner"),
		client:    client,
	}
}

func (c Runner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	natsMembers := []string{}
	for _, addr := range strings.Split(c.addresses, ",") {
		uri := url.URL{
			Scheme: "nats",
			User:   url.UserPassword(c.username, c.password),
			Host:   addr,
		}
		natsMembers = append(natsMembers, uri.String())
	}

	conn, err := yagnats.Connect(natsMembers)
	for err != nil {
		c.logger.Error("connecting-to-nats-failed", err)
		select {
		case <-signals:
			return nil
		case <-time.After(time.Second):
			conn, err = yagnats.Connect(natsMembers)
		}
	}

	*c.client = conn
	c.logger.Info("connecting-to-nats-succeeeded")
	close(ready)

	<-signals
	conn.Close()
	return nil
}
