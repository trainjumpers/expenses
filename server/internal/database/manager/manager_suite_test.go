package database_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestManagerIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DatabaseManager Suite")
}
