package broker

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBrokerGo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BrokerGo Suite")
}
