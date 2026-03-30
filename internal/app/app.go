package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	httpapp "github.com/YagorX/shop-proxy/internal/app/http"
	tcpserv "github.com/YagorX/shop-proxy/internal/app/tcp"
	"github.com/YagorX/shop-proxy/internal/config"
	"github.com/YagorX/shop-proxy/internal/observability"
	httptransport "github.com/YagorX/shop-proxy/internal/transport/http"
)

type App struct {
	logger    *slog.Logger
	httpApp   *httpapp.Server
	tcpServer *tcpserv.Server
	errCh     chan error
}

func New(cfg *config.Config) (*App, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	runtimeLogger := observability.NewLogger(observability.LoggerOptions{
		Service: cfg.ServiceName,
		Env:     cfg.Env,
		Version: cfg.Version,
		Level:   cfg.LogLevel,
	})
	observability.SetDefaultLogger(runtimeLogger.Logger)

	metrics := observability.MustMetrics()

	tcpRuntime, err := tcpserv.New(cfg.Proxy.ListenAddr, cfg.Proxy.UpstreamAddr, runtimeLogger.Logger, metrics)
	if err != nil {
		return nil, fmt.Errorf("create tcp app: %w", err)
	}

	httpRouter := httptransport.NewRouter(
		httptransport.RouterDeps{
			Logger:          runtimeLogger.Logger,
			FaultController: tcpRuntime,
		},
	)

	httpRuntime, err := httpapp.New(cfg.HTTP.Addr, httpRouter, cfg.HTTP.Timeout, runtimeLogger.Logger)
	if err != nil {
		return nil, fmt.Errorf("create http app: %w", err)
	}

	return &App{
		logger:    runtimeLogger.Logger,
		httpApp:   httpRuntime,
		tcpServer: tcpRuntime,

		errCh: make(chan error, 2),
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	if a == nil {
		return errors.New("app is nil")
	}

	go func() {
		if err := a.httpApp.Run(ctx); err != nil {
			a.errCh <- err
			a.logger.Error("http app failed", slog.String("error", err.Error()))
		}
	}()

	go func() {
		if err := a.tcpServer.Run(ctx); err != nil {
			a.errCh <- err
			a.logger.Error("tcp app failed", slog.String("error", err.Error()))
		}
	}()

	a.logger.Info("proxy service bootstrap completed",
		slog.String("http_addr", a.httpApp.Addr()),
		slog.String("tcp_addr", a.tcpServer.Addr()),
	)

	select {
	case <-ctx.Done():
		a.logger.Info("proxy service stopped by context")
		return a.Shutdown(context.Background())
	case err := <-a.errCh:
		if err != nil {
			_ = a.Shutdown(context.Background())
			return err
		}
		return nil
	}
}

func (a *App) Errors() <-chan error {
	return a.errCh
}

func (a *App) Shutdown(ctx context.Context) error {
	if a == nil {
		return nil
	}

	var shutdownErr error

	if a.httpApp != nil {
		if err := a.httpApp.Shutdown(ctx); err != nil {
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("stop http app: %w", err))
		}
	}

	if a.tcpServer != nil {
		if err := a.tcpServer.Shutdown(ctx); err != nil {
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("stop tcp app: %w", err))
		}
	}

	a.logger.Info("proxy service stopped")

	return shutdownErr
}
