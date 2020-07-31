package sse

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSSE(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SSE Consumer Suite")
}
