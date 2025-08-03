package postgres_test

import (
	"time"

	"expenses/pkg/database/manager/postgres"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Connection Pool Configuration", func() {

	Describe("OptimizedPoolConfig", func() {
		It("should return a configuration suitable for production", func() {
			config := postgres.OptimizedPoolConfig()

			Expect(config).NotTo(BeNil())
			Expect(config.MaxConns).To(BeEquivalentTo(25))
			Expect(config.MinConns).To(BeEquivalentTo(5))
			Expect(config.MaxConnLifetime).To(Equal(time.Hour))
			Expect(config.MaxConnIdleTime).To(Equal(30 * time.Minute))
			Expect(config.HealthCheckPeriod).To(Equal(5 * time.Minute))
			Expect(config.ConnectTimeout).To(Equal(10 * time.Second))
			Expect(config.ApplicationName).To(Equal("neurospend"))
			Expect(config.PreferSimpleProtocol).To(BeFalse())
			Expect(config.StatementCacheCapacity).To(Equal(512))
		})
	})

	Describe("DevelopmentPoolConfig", func() {
		It("should return a configuration suitable for development", func() {
			config := postgres.DevelopmentPoolConfig()

			Expect(config).NotTo(BeNil())
			Expect(config.MaxConns).To(BeEquivalentTo(10))
			Expect(config.MinConns).To(BeEquivalentTo(2))
			Expect(config.MaxConnIdleTime).To(Equal(10 * time.Minute))
			Expect(config.ApplicationName).To(Equal("neurospend-dev"))
			Expect(config.PreferSimpleProtocol).To(BeTrue())
			Expect(config.StatementCacheCapacity).To(Equal(128))
		})
	})

})
