package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	addr    string
	timeout time.Duration
	logger  *slog.Logger
	server  *http.Server
}

func New(addr string, handler http.Handler, timeout time.Duration, logger *slog.Logger) (*Server, error) {
	if addr == "" {
		return nil, fmt.Errorf("http address is required")
	}
	if handler == nil {
		return nil, fmt.Errorf("http handler is required")
	}
	if timeout <= 0 {
		return nil, fmt.Errorf("http timeout must be > 0")
	}
	if logger == nil {
		logger = slog.Default()
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return &Server{
		logger:  logger,
		server:  server,
		addr:    addr,
		timeout: timeout,
	}, nil
}

func (a *Server) Run(ctx context.Context) error {
	const op = "httpapp.Run"

	log := a.logger.With(
		slog.String("op", op),
		slog.String("addr", a.server.Addr),
	)

	log.Info("http server started")

	go func() {
		<-ctx.Done()
		_ = a.Shutdown(context.Background())
	}()

	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("http server stopped")

	return nil
}

func (a *Server) Shutdown(ctx context.Context) error {
	const op = "httpapp.Shutdown"

	a.logger.With(
		slog.String("op", op),
		slog.String("addr", a.server.Addr),
	).Info("stopping http server")

	shutdownCtx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *Server) Addr() string {
	if a == nil || a.server == nil {
		return ""
	}
	return a.server.Addr
}
