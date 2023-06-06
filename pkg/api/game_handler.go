package api

import (
	"encoding/json"
	"fmt"
	"hive/pkg/game"
	"net/http"
	"sync"

	"go.uber.org/zap"
)

type GameHandler struct {
	log     *zap.Logger
	service Service
}

func NewBuildService(l *zap.Logger, s Service) *GameHandler {
	return &GameHandler{log: l, service: s}
}

func (h *GameHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/start", h.handleStart)
	mux.HandleFunc("/signal", h.handleSignal)
}

func (h *GameHandler) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	var req GameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	sw := &statusWriter{
		writerAlive: make(chan struct{}),
		encoder:     json.NewEncoder(w),
		rc:          http.NewResponseController(w),
	}

	if err := h.service.StartGame(ctx, &req, sw); err != nil {
		h.log.Error("Failed to start build", zap.Error(err))

		w.Header().Set("X-Error-Message", fmt.Sprintf("Failed to signal build: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		err = sw.Updated(&StatusUpdate{})
		if err != nil {
			h.log.Error("Failed to start build", zap.Error(err))
			return
		}
		return
	}
	select {
	case <-ctx.Done():
	case <-sw.writerAlive:
	}

	sw.m.Lock()
	defer sw.m.Unlock()

}

func (h *GameHandler) handleSignal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var gameID game.ID
	if err := gameID.UnmarshalText([]byte(r.URL.Query().Get("build_id"))); err != nil {
		http.Error(w, fmt.Sprintf("Invalid build_id: %v", err), http.StatusBadRequest)
		return
	}

	var signalRequest SignalRequest
	if err := json.NewDecoder(r.Body).Decode(&signalRequest); err != nil {
		http.Error(w, fmt.Sprintf("Failed to decode signal request: %v", err), http.StatusBadRequest)
		return
	}

	signalResponse, err := h.service.SignalGame(r.Context(), gameID, &signalRequest)
	if err != nil {
		h.log.Error("Failed to signal build", zap.Error(err))

		w.Header().Set("X-Error-Message", fmt.Sprintf("Failed to signal build: %v", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(signalResponse); err != nil {
		h.log.Error("Failed to encode signal response", zap.Error(err))
	}

}

type statusWriter struct {
	m           sync.Mutex
	writerAlive chan struct{}
	encoder     *json.Encoder
	rc          *http.ResponseController
}

func (sw *statusWriter) Started(rsp *GameStarted) error {
	sw.m.Lock()
	defer sw.m.Unlock()

	err := sw.encoder.Encode(rsp)
	if err != nil {
		return err
	}
	return sw.rc.Flush()
}

func (sw *statusWriter) Updated(update *StatusUpdate) error {
	sw.m.Lock()
	defer sw.m.Unlock()

	err := sw.encoder.Encode(update)
	if err != nil {
		return err
	}
	return sw.rc.Flush()
}
