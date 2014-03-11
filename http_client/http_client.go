package http_client

import (
	"crypto/tls"
	"net/http"
)

func New(skipSSLVerification bool) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipSSLVerification,
		},
	}

	return &http.Client{
		Transport: transport,
	}
}
