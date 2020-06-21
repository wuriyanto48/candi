package userservice

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/config"
	"agungdwiprasetyo.com/backend-microservices/config/broker"
	"agungdwiprasetyo.com/backend-microservices/config/database"
	"agungdwiprasetyo.com/backend-microservices/internal/user-service/modules/auth"
	"agungdwiprasetyo.com/backend-microservices/internal/user-service/modules/customer"
	"agungdwiprasetyo.com/backend-microservices/internal/user-service/modules/member"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/dependency"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"agungdwiprasetyo.com/backend-microservices/pkg/middleware"
	authsdk "agungdwiprasetyo.com/backend-microservices/pkg/sdk/auth-service"
)

// Service model
type Service struct {
	deps    dependency.Dependency
	modules []factory.ModuleFactory
	name    constant.Service
}

// NewService starting service
func NewService(serviceName string, cfg *config.Config) factory.ServiceFactory {
	var depsOptions = []dependency.Option{
		dependency.SetMiddleware(middleware.NewMiddleware(authsdk.NewAuthServiceGRPC())),
	}

	cfg.Load(
		func(ctx context.Context) interfaces.Closer {
			d := database.InitMongoDB(ctx)
			depsOptions = append(depsOptions, dependency.SetMongoDatabase(d))
			return d
		},
		func(context.Context) interfaces.Closer {
			d := database.InitRedis()
			depsOptions = append(depsOptions, dependency.SetRedisPool(d))
			return d
		},
		func(context.Context) interfaces.Closer {
			d := database.InitSQLDatabase()
			depsOptions = append(depsOptions, dependency.SetSQLDatabase(d))
			return d
		},
		func(context.Context) interfaces.Closer {
			d := broker.InitKafkaBroker(config.BaseEnv().Kafka.ClientID)
			depsOptions = append(depsOptions, dependency.SetBroker(d))
			return d
		},
	)

	// init all service dependencies
	deps := dependency.InitDependency(depsOptions...)

	modules := []factory.ModuleFactory{
		member.NewModule(deps),
		customer.NewModule(deps),
		auth.NewModule(deps),
	}

	return &Service{
		deps:    deps,
		modules: modules,
		name:    constant.Service(serviceName),
	}
}

// GetDependency method
func (s *Service) GetDependency() dependency.Dependency {
	return s.deps
}

// GetModules method
func (s *Service) GetModules() []factory.ModuleFactory {
	return s.modules
}

// Name method
func (s *Service) Name() constant.Service {
	return s.name
}
