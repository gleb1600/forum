package mmrstorage_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMemoryStorage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MemoryStorage Suite")
}
