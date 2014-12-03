package faketimeprovider_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestFaketimeprovider(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Faketimeprovider Suite")
}
