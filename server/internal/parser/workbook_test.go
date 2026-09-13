package parser

import (
	"expenses/pkg/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ValidateWorkbookSize", func() {
	It("accepts a normal workbook", func() {
		xlsx := utils.CreateXLSXFile([][]string{{"a", "b"}, {"1", "2"}})
		Expect(ValidateWorkbookSize(xlsx)).To(Succeed())
	})

	It("rejects input that is not a zip archive", func() {
		Expect(ValidateWorkbookSize([]byte("not a zip"))).NotTo(Succeed())
	})

	It("rejects a workbook whose declared size exceeds the limit", func() {
		xlsx := utils.CreateXLSXFile([][]string{{"a", "b"}, {"1", "2"}})
		Expect(validateWorkbookSize(xlsx, 1)).NotTo(Succeed())
	})
})
