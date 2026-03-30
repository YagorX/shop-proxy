package http

import (
	"log/slog"
	"net/http"

	"github.com/YagorX/shop-proxy/internal/transport/http/contracts"
	"github.com/YagorX/shop-proxy/internal/transport/http/handlers"
	"github.com/YagorX/shop-proxy/internal/transport/http/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type RouterDeps struct {
	Logger          *slog.Logger
	FaultController contracts.FaultController
	AuthService     contracts.AuthService
}

func NewRouter(router RouterDeps) http.Handler {
	logger := router.Logger
	if logger == nil {
		logger = slog.Default()
	}

	mux := http.NewServeMux()
	adminMiddleware := middleware.AdminOnly(logger, router.AuthService)

	healthHandler := handlers.NewHealthHandler()
	adminHandler := handlers.NewAdminHandler(router.FaultController)

	mux.HandleFunc("/health", healthHandler.Health)
	mux.HandleFunc("/ready", healthHandler.Ready)
	mux.Handle("/metrics", promhttp.Handler())

	mux.Handle("/admin/state", adminMiddleware(http.HandlerFunc(adminHandler.State)))
	mux.Handle("/admin/faults/delay", adminMiddleware(http.HandlerFunc(adminHandler.SetDelay)))
	mux.Handle("/admin/reset", adminMiddleware(http.HandlerFunc(adminHandler.Reset)))

	return mux
}
