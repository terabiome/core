package singletons

import (
	"errors"
	"fmt"
	"log"

	"github.com/terabiome/core/internal/config"
	"github.com/terabiome/core/internal/database"
	"github.com/terabiome/core/internal/repositories/postgresql"
	"github.com/terabiome/core/internal/services"
	"github.com/terabiome/core/pkg/logging"
	"github.com/terabiome/core/pkg/messaging/natsjetstream"
)

type Singletons struct {
	TxManager *postgresql.TxManager
	messagingSingletons
	repositorySingletons
	serviceSingletons
}

type repositorySingletons struct {
	UnitOfWorkRepository *postgresql.UnitOfWorkPostgresRepository
	DispatchRepository   *postgresql.DispatchPostgresRepository
}

type serviceSingletons struct {
	UnitOfWorkService *services.UnitOfWork
	DispatchService   *services.Dispatch
}

type messagingSingletons struct {
	NATSJetstream *natsjetstream.Client
}

var sings *Singletons

func InitLogger() {
	cfg := config.Get().Logger
	logging.InitGlobalSlogger(logging.Config{
		Level:  cfg.Level,
		Format: cfg.Format,
	})
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
	repositoryWrapper := repositorySingletons{
		UnitOfWorkRepository: postgresql.NewUnitOfWorkPostgresRepository(pool),
		DispatchRepository:   postgresql.NewDispatchPostgresRepository(pool),
	}

	sings = &Singletons{
		TxManager:            txManager,
		repositorySingletons: repositoryWrapper,
		UnitOfWorkService: services.NewUnitOfWork(
			txManager,
			repositoryWrapper.UnitOfWorkRepository,
		),
	}

	logging.GetGlobalSlogger().Info("Constructed singletons")
	return nil
}

// InitDispatch wires NATS and the dispatch batch service. Call after Init.
func InitDispatch() error {
	if sings == nil || sings.TxManager == nil || sings.UnitOfWorkRepository == nil || sings.DispatchRepository == nil {
		return errors.New("singletons.Init must run before InitDispatch")
	}

	js, err := natsjetstream.New(
		natsjetstream.Config{
			URL: config.Get().NATS.URL,
		},
		logging.NewSlogger(
			logging.Config{
				Level:  config.Get().Logger.Level,
				Format: config.Get().Logger.Format,
			},
			map[string]any{
				"scope": "jetstream",
			},
		),
	)
	if err != nil {
		return fmt.Errorf("failed to setup NATS Jetstream wrapper: %w", err)
	}

	sings.messagingSingletons = messagingSingletons{NATSJetstream: js}
	sings.DispatchService = services.NewDispatch(
		sings.TxManager,
		sings.UnitOfWorkRepository,
		sings.DispatchRepository,
		js,
	)
	return nil
}

func CloseDispatch() error {
	if sings == nil || sings.NATSJetstream == nil {
		return nil
	}
	if err := sings.NATSJetstream.Close(); err != nil {
		return fmt.Errorf("failed to close NATS connection: %w", err)
	}
	return nil
}

func Get() *Singletons {
	return sings
}
