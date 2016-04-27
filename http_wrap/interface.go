/*
Package http wraps golang http in an interface.
 */
package http_wrap

import os_http "net/http"

//go:generate counterfeiter -o httpfakes/fake_http_client.go . Client

type Client interface {
	Do(req *os_http.Request) (resp *os_http.Response, err error)
}