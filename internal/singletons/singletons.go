package singletons

import (
	"errors"
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

func InitLogger() {
	logging.InitGlobalSlogger(config.Get().Logger)
}

func Init() error {
	log.Println("Constructing singletons")
	InitLogger()

	if err := database.Init(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// hardcode postgresql for now
	pool := database.GetPool()
	txManager := postgresql.NewTxManager(pool)
	repositoryWrapper := RepositoryWrapper{
		UnitOfWorkRepository: postgresql.NewUnitOfWorkPostgresRepository(pool),
		DispatchRepository:   postgresql.NewDispatchPostgresRepository(pool),
	}

	sings = &Singletons{
		TxManager:         txManager,
		RepositoryWrapper: repositoryWrapper,
		UnitOfWorkService: services.NewUnitOfWork(
			txManager,
			repositoryWrapper.UnitOfWorkRepository,
		),
	}

	logging.GetGlobalSlogger().Info("Constructed singletons")
	return nil
}

// InitDispatch wires NATS and the dispatch batch service. Call after Init.
// Do not call this from the API server until a worker loop in that process actually sweeps dispatches.
func InitDispatch() error {
	if sings == nil || sings.TxManager == nil || sings.UnitOfWorkRepository == nil || sings.DispatchRepository == nil {
		return errors.New("singletons.Init must run before InitDispatch")
	}

	js, err := nats.NewNATSJetstream()
	if err != nil {
		return fmt.Errorf("failed to setup NATS Jetstream wrapper: %w", err)
	}

	sings.MessagingWrapper = MessagingWrapper{NATSJetstream: js}
	sings.DispatchService = services.NewDispatch(
		sings.TxManager,
		sings.UnitOfWorkRepository,
		sings.DispatchRepository,
		js,
	)
	return nil
}

func Get() *Singletons {
	return sings
}
