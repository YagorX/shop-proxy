package http

import (
	"log/slog"
	"net/http"

	"github.com/YagorX/shop-proxy/internal/transport/http/contracts"
	"github.com/YagorX/shop-proxy/internal/transport/http/handlers"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type RouterDeps struct {
	Logger          *slog.Logger
	FaultController contracts.FaultController
}

func NewRouter(router RouterDeps) http.Handler {
	mux := http.NewServeMux()

	healthHandler := handlers.NewHealthHandler()
	adminHandler := handlers.NewAdminHandler(router.FaultController)

	mux.HandleFunc("/health", healthHandler.Health)
	mux.HandleFunc("/ready", healthHandler.Ready)
	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/admin/state", adminHandler.State)
	mux.HandleFunc("/admin/faults/delay", adminHandler.SetDelay)
	mux.HandleFunc("/admin/reset", adminHandler.Reset)

	mux.Handle("/metrics", promhttp.Handler())

	return mux
}
