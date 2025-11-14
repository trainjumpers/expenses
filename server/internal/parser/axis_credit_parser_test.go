package parser_test

import (
	"expenses/internal/models"
	"expenses/internal/parser"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AxisCreditParser", func() {
	It("should be registered", func() {
		_, ok := parser.GetParser(models.BankTypeAxisCredit)
		Expect(ok).To(BeTrue())
	})
})
