//go:build wireinject
// +build wireinject

package wire

import (
	"expenses/internal/api"
	"expenses/internal/api/controller"
	"expenses/internal/config"
	database "expenses/internal/database/manager"
	"expenses/internal/repository"
	"expenses/internal/service"
	"expenses/internal/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

type Provider struct {
	Handler   *gin.Engine
	dbManager database.DatabaseManager
}

// Close all connections app makes in various places
func (p *Provider) Close() error {
	return p.dbManager.Close()
}

func NewProvider(handler *gin.Engine, dbManager database.DatabaseManager) *Provider {
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
	database.NewDatabaseManager,
	config.NewConfig,
	api.Init,
	controllerSet,
	repositorySet,
	serviceSet,
	validatorSet,
)

var controllerSet = wire.NewSet(
	controller.NewAuthController,
	controller.NewStatementController,
)

var repositorySet = wire.NewSet(
	repository.NewUserRepository,
	repository.NewAccountRepository,
	repository.NewCategoryRepository,
	repository.NewTransactionRepository,
	repository.NewRuleRepository,
	repository.NewStatementRepository,
)

var serviceSet = wire.NewSet(
	service.NewUserService,
	service.NewAuthService,
	service.NewAccountService,
	service.NewCategoryService,
	service.NewTransactionService,
	service.NewRuleService,
	service.NewStatementService,
)

var validatorSet = wire.NewSet(
	validator.NewStatementValidator,
)
