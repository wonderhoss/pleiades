package sse

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/op/go-logging"
)

func TestSSE(t *testing.T) {
	logging.InitForTesting(logging.CRITICAL)
	RegisterFailHandler(Fail)
	RunSpecs(t, "SSE Consumer Suite")
}
