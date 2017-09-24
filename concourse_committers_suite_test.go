package concourse_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestConcourseCommitters(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ConcourseCommitters Suite")
}
