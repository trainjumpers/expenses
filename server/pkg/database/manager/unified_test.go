package manager_test

import (
	"testing"

	"expenses/internal/config"
	"expenses/pkg/database/manager"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUnifiedManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unified Database Manager Suite")
}

var _ = Describe("Unified Database Manager", func() {
	var cfg *config.Config

	BeforeEach(func() {
		cfg = &config.Config{
			DBSchema: "public",
		}
	})

	Describe("Configuration", func() {
		It("should create manager with default config", func() {
			dbManager, err := manager.NewDatabaseManager(cfg)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbManager).NotTo(BeNil())

			// Check that it has all features
			config := dbManager.GetConfig()
			Expect(config.EnableTransactions).To(BeTrue())
			Expect(config.EnableLocks).To(BeTrue())

			dbManager.Close()
		})

		It("should create manager with basic config", func() {
			dbManager, err := manager.NewBasicDatabaseManager(cfg)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbManager).NotTo(BeNil())

			// Check that advanced features are disabled
			config := dbManager.GetConfig()
			Expect(config.EnableTransactions).To(BeTrue())
			Expect(config.EnableLocks).To(BeTrue())
			Expect(config.EnableRetry).To(BeFalse())
			Expect(config.EnableMonitoring).To(BeFalse())

			dbManager.Close()
		})

		It("should create manager with production config", func() {
			dbManager, err := manager.NewProductionDatabaseManager(cfg)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbManager).NotTo(BeNil())

			// Check that all features are enabled
			config := dbManager.GetConfig()
			Expect(config.EnableTransactions).To(BeTrue())
			Expect(config.EnableLocks).To(BeTrue())
			Expect(config.EnableRetry).To(BeTrue())
			Expect(config.EnableSavepoints).To(BeTrue())
			Expect(config.EnableBatch).To(BeTrue())
			Expect(config.EnableMonitoring).To(BeTrue())

			dbManager.Close()
		})

		It("should create manager with custom config", func() {
			customConfig := &manager.DatabaseManagerConfig{
				EnableTransactions: true,
				EnableLocks:        true,
				EnableRetry:        true,
				EnableSavepoints:   false, // Disabled
				EnableBatch:        true,
				EnableMonitoring:   false, // Disabled
			}

			dbManager, err := manager.NewDatabaseManagerWithConfig(cfg, customConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbManager).NotTo(BeNil())

			// Check custom configuration
			config := dbManager.GetConfig()
			Expect(config.EnableRetry).To(BeTrue())
			Expect(config.EnableSavepoints).To(BeFalse())
			Expect(config.EnableMonitoring).To(BeFalse())

			dbManager.Close()
		})
	})

	Describe("Feature Detection", func() {
		It("should detect enabled features correctly", func() {
			dbManager, err := manager.NewProductionDatabaseManager(cfg)
			Expect(err).NotTo(HaveOccurred())
			defer dbManager.Close()

			// Check feature detection
			Expect(dbManager.IsFeatureEnabled(manager.FeatureRetry)).To(BeTrue())
			Expect(dbManager.IsFeatureEnabled(manager.FeatureSavepoints)).To(BeTrue())
			Expect(dbManager.IsFeatureEnabled(manager.FeatureBatch)).To(BeTrue())
			Expect(dbManager.IsFeatureEnabled(manager.FeatureMonitoring)).To(BeTrue())
		})

		It("should detect disabled features correctly", func() {
			dbManager, err := manager.NewBasicDatabaseManager(cfg)
			Expect(err).NotTo(HaveOccurred())
			defer dbManager.Close()

			// Check feature detection
			Expect(dbManager.IsFeatureEnabled(manager.FeatureRetry)).To(BeFalse())
			Expect(dbManager.IsFeatureEnabled(manager.FeatureSavepoints)).To(BeFalse())
			Expect(dbManager.IsFeatureEnabled(manager.FeatureBatch)).To(BeFalse())
			Expect(dbManager.IsFeatureEnabled(manager.FeatureMonitoring)).To(BeFalse())
		})
	})
})
