package http_client

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

func New(skipSSLVerification bool, timeout time.Duration) *http.Client {
	dialFunc := func(network, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(network, addr, timeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(timeout))
		return conn, err
	}

	transport := &http.Transport{
		Dial: dialFunc,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipSSLVerification,
		},
	}

	return &http.Client{
		Transport: transport,
	}
}
