package errors

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestErrors(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Errors Suite")
}

var _ = Describe("Error Formatting", func() {
	Describe("formatError", func() {
		var (
			originalErr error
			message     string
			errorType   string
			status      int
			err         *AuthError
		)

		BeforeEach(func() {
			originalErr = errors.New("test error")
			message = "test message"
			errorType = "TestError"
			status = http.StatusBadRequest
			err = formatError(status, message, originalErr, errorType)
		})

		It("should set the correct message", func() {
			Expect(err.Message).To(Equal(message))
		})

		It("should set the correct error type", func() {
			Expect(err.ErrorType).To(Equal(errorType))
		})

		It("should set the correct status code", func() {
			Expect(err.Status).To(Equal(status))
		})

		It("should set the original error", func() {
			Expect(err.Err).To(Equal(originalErr))
		})

		It("should have a non-empty stack trace", func() {
			Expect(err.Stack).NotTo(BeEmpty())
		})

		It("should have valid stack trace lines", func() {
			for _, line := range err.Stack {
				Expect(strings.TrimSpace(line)).NotTo(BeEmpty())
				Expect(line).To(ContainSubstring("."))
			}
		})

		It("should return the message when Error() is called", func() {
			Expect(err.Error()).To(Equal(message))
		})

		It("should return the original error when Unwrap() is called", func() {
			Expect(err.Unwrap()).To(Equal(originalErr))
		})
	})
})
