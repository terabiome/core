package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	apihttp "github.com/terabiome/core/internal/api"
	"github.com/terabiome/core/internal/config"
	"github.com/terabiome/core/internal/database"
	httphandlers "github.com/terabiome/core/internal/handlers/http"
	"github.com/terabiome/core/internal/singletons"
	"github.com/terabiome/core/pkg/logging"
)

const apiSubcommand subcommand = "api"

type apiCLIInput struct {
	Addr string
}

func processAPISubcommand(ctx context.Context, apiInput apiCLIInput) error {
	if err := singletons.Init(); err != nil {
		return fmt.Errorf("failed to initialize singletons: %w", err)
	}
	defer database.Close()

	if err := singletons.InitDispatch(); err != nil {
		return fmt.Errorf("failed to initialize dispatch: %w", err)
	}
	defer func() {
		if err := singletons.CloseDispatch(); err != nil {
			logging.GetGlobalSlogger().Error("failed to close NATS", "error", err)
		}
	}()

	dispatchCfg := config.Get().Dispatch
	sweepCtx, stopSweep := context.WithCancel(ctx)
	defer stopSweep()
	sweepDone := make(chan struct{})
	go func() {
		defer close(sweepDone)
		singletons.Get().DispatchService.RunSweep(
			sweepCtx,
			dispatchCfg.SweepLimit,
			dispatchCfg.SweepInterval,
			dispatchCfg.SweepPassTimeout,
		)
	}()

	server := &http.Server{
		Addr:              apiInput.Addr,
		Handler:           apihttp.SetupMux(httphandlers.NewUnitOfWork(singletons.Get().UnitOfWorkService)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logging.GetGlobalSlogger().Info("api listening", "addr", apiInput.Addr)
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	var serveErr error
	select {
	case <-ctx.Done():
	case serveErr = <-errCh:
	}

	// Stop after the current pass. Do not drain HTTP first, and do not block ListenAndServe on this loop.
	stopSweep()
	select {
	case <-sweepDone:
	case <-time.After(30 * time.Second):
		logging.GetGlobalSlogger().Error("dispatch sweep did not stop before timeout")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return errors.Join(serveErr, err)
	}
	return serveErr
}
