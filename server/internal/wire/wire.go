//go:build wireinject
// +build wireinject

package wire

import (
	"expenses/internal/api"
	"expenses/internal/api/controller"
	"expenses/internal/config"
	"expenses/internal/repository"
	"expenses/internal/service"
	"expenses/internal/validator"
	"expenses/pkg/database/manager"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

type Provider struct {
	Handler   *gin.Engine
	dbManager manager.DatabaseManager
}

// Close all connections app makes in various places
func (p *Provider) Close() error {
	return p.dbManager.Close()
}

func NewProvider(handler *gin.Engine, dbManager manager.DatabaseManager) *Provider {
	return &Provider{
		Handler:   handler,
		dbManager: dbManager,
	}
}

func InitializeApplication() (*Provider, error) {
	wire.Build(ProviderSet)
	return &Provider{}, nil
}

var ProviderSet = wire.NewSet(
	NewProvider,
	manager.NewDatabaseManager,
	config.NewConfig,
	api.Init,
	controllerSet,
	repositorySet,
	serviceSet,
	validatorSet,
)

var controllerSet = wire.NewSet(
	controller.NewAccountController,
	controller.NewAuthController,
	controller.NewCategoryController,
	controller.NewRuleController,
	controller.NewStatementController,
	controller.NewTransactionController,
)

var repositorySet = wire.NewSet(
	repository.NewAccountRepository,
	repository.NewCategoryRepository,
	repository.NewRuleRepository,
	repository.NewStatementRepository,
	repository.NewTransactionRepository,
	repository.NewUserRepository,
)

var serviceSet = wire.NewSet(
	service.NewAccountService,
	service.NewAuthService,
	service.NewCategoryService,
	service.NewRuleEngineService,
	service.NewRuleService,
	service.NewStatementService,
	service.NewTransactionService,
	service.NewUserService,
)

var validatorSet = wire.NewSet(
	validator.NewStatementValidator,
)
