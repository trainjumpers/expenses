package base_test

import (
	"os"
	"time"

	"expenses/pkg/database/manager/base"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DatabaseManagerConfig", func() {

	Describe("DefaultConfig", func() {
		It("should return a production-ready configuration", func() {
			config := base.DefaultConfig()

			Expect(config.EnableTransactions).To(BeTrue())
			Expect(config.EnableLocks).To(BeTrue())
			Expect(config.EnableRetry).To(BeTrue())
			Expect(config.EnableSavepoints).To(BeTrue())
			Expect(config.EnableBatch).To(BeTrue())
			Expect(config.DefaultTimeout).To(Equal(30 * time.Second))
			Expect(config.EnableMonitoring).To(BeTrue())
			Expect(config.EnableMetrics).To(BeTrue())
			Expect(config.OptimizePool).To(BeTrue())
		})
	})

	Describe("BasicConfig", func() {
		It("should return a minimal configuration", func() {
			config := base.BasicConfig()

			Expect(config.EnableTransactions).To(BeTrue())
			Expect(config.EnableLocks).To(BeTrue())
			Expect(config.EnableRetry).To(BeFalse())
			Expect(config.EnableSavepoints).To(BeFalse())
			Expect(config.EnableBatch).To(BeFalse())
			Expect(config.DefaultTimeout).To(Equal(10 * time.Second))
			Expect(config.EnableMonitoring).To(BeFalse())
			Expect(config.EnableMetrics).To(BeFalse())
			Expect(config.OptimizePool).To(BeFalse())
		})
	})

	Describe("DevelopmentConfig", func() {
		It("should return a development-optimized configuration", func() {
			config := base.DevelopmentConfig()

			Expect(config.EnableTransactions).To(BeTrue())
			Expect(config.EnableLocks).To(BeTrue())
			Expect(config.EnableRetry).To(BeFalse())
			Expect(config.EnableSavepoints).To(BeTrue())
			Expect(config.EnableBatch).To(BeTrue())
			Expect(config.DefaultTimeout).To(Equal(5 * time.Second))
			Expect(config.EnableMonitoring).To(BeTrue())
			Expect(config.EnableMetrics).To(BeTrue())
			Expect(config.OptimizePool).To(BeTrue())
		})
	})

	Describe("AutoConfig", func() {
		var originalGinMode string

		BeforeEach(func() {
			originalGinMode = os.Getenv("GIN_MODE")
		})

		AfterEach(func() {
			os.Setenv("GIN_MODE", originalGinMode)
		})

		Context("when GIN_MODE is 'debug'", func() {
			It("should return the development configuration", func() {
				os.Setenv("GIN_MODE", "debug")
				config := base.AutoConfig()
				devConfig := base.DevelopmentConfig()
				Expect(config).To(Equal(devConfig))
			})
		})

		Context("when GIN_MODE is not 'debug'", func() {
			It("should return the default configuration", func() {
				os.Setenv("GIN_MODE", "release")
				config := base.AutoConfig()
				defaultConfig := base.DefaultConfig()
				Expect(config).To(Equal(defaultConfig))
			})
		})

		Context("when GIN_MODE is not set", func() {
			It("should return the default configuration", func() {
				os.Unsetenv("GIN_MODE")
				config := base.AutoConfig()
				defaultConfig := base.DefaultConfig()
				Expect(config).To(Equal(defaultConfig))
			})
		})
	})
})
