package http_client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestHttp_client(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Http_client Suite")
}
