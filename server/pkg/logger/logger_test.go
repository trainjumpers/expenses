package logger_test

import (
	"os"
	"testing"

	"expenses/pkg/logger"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

func TestLogger(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logger Suite")
}

var _ = Describe("Logger", func() {
	Describe("getLoggingLevel", func() {
		AfterEach(func() {
			os.Unsetenv("LOGGING_LEVEL")
			os.Unsetenv("ENV")
		})

		It("should return debug level for LOGGING_LEVEL=debug", func() {
			orig := os.Getenv("LOGGING_LEVEL")
			os.Setenv("LOGGING_LEVEL", "debug")
			defer os.Setenv("LOGGING_LEVEL", orig)
			level := loggerTestGetLoggingLevel()
			Expect(level.Level()).To(Equal(zap.DebugLevel))
		})
		It("should return info level for LOGGING_LEVEL=info", func() {
			orig := os.Getenv("LOGGING_LEVEL")
			os.Setenv("LOGGING_LEVEL", "info")
			defer os.Setenv("LOGGING_LEVEL", orig)
			level := loggerTestGetLoggingLevel()
			Expect(level.Level()).To(Equal(zap.InfoLevel))
		})
		It("should return warn level for LOGGING_LEVEL=warn", func() {
			orig := os.Getenv("LOGGING_LEVEL")
			os.Setenv("LOGGING_LEVEL", "warn")
			defer os.Setenv("LOGGING_LEVEL", orig)
			level := loggerTestGetLoggingLevel()
			Expect(level.Level()).To(Equal(zap.WarnLevel))
		})
		It("should return error level for LOGGING_LEVEL=error", func() {
			orig := os.Getenv("LOGGING_LEVEL")
			os.Setenv("LOGGING_LEVEL", "error")
			defer os.Setenv("LOGGING_LEVEL", orig)
			level := loggerTestGetLoggingLevel()
			Expect(level.Level()).To(Equal(zap.ErrorLevel))
		})
		It("should return panic level for LOGGING_LEVEL=panic", func() {
			orig := os.Getenv("LOGGING_LEVEL")
			os.Setenv("LOGGING_LEVEL", "panic")
			defer os.Setenv("LOGGING_LEVEL", orig)
			level := loggerTestGetLoggingLevel()
			Expect(level.Level()).To(Equal(zap.PanicLevel))
		})
		It("should return fatal level for LOGGING_LEVEL=fatal", func() {
			orig := os.Getenv("LOGGING_LEVEL")
			os.Setenv("LOGGING_LEVEL", "fatal")
			defer os.Setenv("LOGGING_LEVEL", orig)
			level := loggerTestGetLoggingLevel()
			Expect(level.Level()).To(Equal(zap.FatalLevel))
		})
		It("should return correct level for config environments", func() {
			orig := os.Getenv("ENV")
			os.Unsetenv("LOGGING_LEVEL")
			os.Setenv("ENV", "prod")
			defer os.Setenv("ENV", orig)
			level := loggerTestGetLoggingLevel()
			Expect(level.Level()).To(Equal(zap.InfoLevel))
			os.Setenv("ENV", "dev")
			level = loggerTestGetLoggingLevel()
			Expect(level.Level()).To(Equal(zap.DebugLevel))
			os.Setenv("ENV", "test")
			level = loggerTestGetLoggingLevel()
			Expect(level.Level()).To(Equal(zap.PanicLevel))
		})
		It("should default to info level for unknown env", func() {
			orig := os.Getenv("ENV")
			os.Unsetenv("LOGGING_LEVEL")
			os.Setenv("ENV", "unknown")
			defer os.Setenv("ENV", orig)
			level := loggerTestGetLoggingLevel()
			Expect(level.Level()).To(Equal(zap.InfoLevel))
		})
	})

	Describe("getSamplingConfig", func() {
		It("should return sampling config in prod", func() {
			orig := os.Getenv("ENV")
			os.Setenv("ENV", "prod")
			defer os.Setenv("ENV", orig)
			cfg := loggerTestGetSamplingConfig()
			Expect(cfg).NotTo(BeNil())
			Expect(cfg.Initial).To(Equal(100))
			Expect(cfg.Thereafter).To(Equal(100))
		})
		It("should return nil in dev", func() {
			orig := os.Getenv("ENV")
			os.Setenv("ENV", "dev")
			defer os.Setenv("ENV", orig)
			cfg := loggerTestGetSamplingConfig()
			Expect(cfg).To(BeNil())
		})
		It("should return nil in test", func() {
			orig := os.Getenv("ENV")
			os.Setenv("ENV", "test")
			defer os.Setenv("ENV", orig)
			cfg := loggerTestGetSamplingConfig()
			Expect(cfg).To(BeNil())
		})
	})
})

// helpers to access unexported functions for testing
func loggerTestGetLoggingLevel() zap.AtomicLevel {
	return loggerTestGetLoggingLevelImpl()
}

func loggerTestGetLoggingLevelImpl() zap.AtomicLevel {
	return logger.GetLoggingLevel()
}

func loggerTestGetSamplingConfig() *zap.SamplingConfig {
	return logger.GetSamplingConfig()
}
