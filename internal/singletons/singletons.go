package singletons

import (
	"fmt"
	"log"

	"github.com/terabiome/core/internal/config"
	"github.com/terabiome/core/internal/database"
	"github.com/terabiome/core/internal/messaging/nats"
	"github.com/terabiome/core/internal/repositories/postgresql"
	"github.com/terabiome/core/internal/services"
	"github.com/terabiome/core/pkg/logging"
)

type Singletons struct {
	TxManager *postgresql.TxManager
	MessagingWrapper
	RepositoryWrapper
	ServiceWrapper
}

type RepositoryWrapper struct {
	UnitOfWorkRepository *postgresql.UnitOfWorkPostgresRepository
	DispatchRepository   *postgresql.DispatchPostgresRepository
}

type ServiceWrapper struct {
	UnitOfWorkService *services.UnitOfWork
	DispatchService   *services.Dispatch
}

type MessagingWrapper struct {
	NATSJetstream *nats.NATSJetstream
}

var sings *Singletons

func Init() error {
	log.Println("Constructing singletons")

	logging.InitGlobalSlogger(config.Get().Logger)

	js, err := nats.NewNATSJetstream()
	if err != nil {
		return fmt.Errorf("failed to setup NATS Jetstream wrapper: %w", err)
	}
	messagingWrapper := MessagingWrapper{
		NATSJetstream: js,
	}

	// hardcode postgresql for now
	pool := database.GetPool()
	txManager := postgresql.NewTxManager(pool)
	repositoryWrapper := RepositoryWrapper{
		UnitOfWorkRepository: postgresql.NewUnitOfWorkPostgresRepository(pool),
		DispatchRepository:   postgresql.NewDispatchPostgresRepository(pool),
	}

	serviceWrapper := ServiceWrapper{
		DispatchService: services.NewDispatch(
			txManager,
			repositoryWrapper.UnitOfWorkRepository,
			repositoryWrapper.DispatchRepository,
			messagingWrapper.NATSJetstream,
		),
	}

	sings = &Singletons{
		TxManager:         txManager,
		MessagingWrapper:  messagingWrapper,
		RepositoryWrapper: repositoryWrapper,
		ServiceWrapper:    serviceWrapper,
	}

	// showing off
	logging.GetGlobalSlogger().Info("Constructed singletons")
	return nil
}

func Get() *Singletons {
	return sings
}
