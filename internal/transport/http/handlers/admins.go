package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/YagorX/shop-proxy/internal/transport/http/contracts"
)

type AdminHandler struct {
	ctrl contracts.FaultController
}

func NewAdminHandler(ctrl contracts.FaultController) *AdminHandler {
	return &AdminHandler{ctrl: ctrl}
}

func (h *AdminHandler) State(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if h == nil || h.ctrl == nil {
		http.Error(w, "fault controller is not configured", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":   "ok",
		"delay_ms": h.ctrl.Delay().Milliseconds(),
	})
}

func (h *AdminHandler) SetDelay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if h == nil || h.ctrl == nil {
		http.Error(w, "fault controller is not configured", http.StatusInternalServerError)
		return
	}

	msRaw := r.URL.Query().Get("ms")
	if msRaw == "" {
		http.Error(w, "ms query param is required", http.StatusBadRequest)
		return
	}

	ms, err := strconv.Atoi(msRaw)
	if err != nil {
		http.Error(w, "ms must be a valid integer", http.StatusBadRequest)
		return
	}
	if ms < 0 {
		http.Error(w, "ms must be >= 0", http.StatusBadRequest)
		return
	}

	delay := time.Duration(ms) * time.Millisecond
	h.ctrl.SetDelay(delay)

	writeJSON(w, http.StatusOK, map[string]any{
		"status":   "ok",
		"delay_ms": delay.Milliseconds(),
	})
}

func (h *AdminHandler) Reset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if h == nil || h.ctrl == nil {
		http.Error(w, "fault controller is not configured", http.StatusInternalServerError)
		return
	}

	h.ctrl.SetDelay(0)

	writeJSON(w, http.StatusOK, map[string]any{
		"status":   "ok",
		"delay_ms": int64(0),
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Default().Error("writeJSON: failed to encode response", slog.String("error", err.Error()))
	}
}
