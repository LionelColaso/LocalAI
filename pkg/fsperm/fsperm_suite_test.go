package fsperm_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFsperm(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LocalAI fsperm test")
}
