package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/JpUnique/TinyGo/pkg/service"
	"github.com/go-chi/chi/v5"
)

// request/response structures
type shortenRequest struct {
	URL         string `json:"url"`
	CustomAlias string `json:"custom_alias,omitempty"`
}

type shortenResponse struct {
	ShortURL string `json:"short_url"`
	ShortKey string `json:"short_key"`
}

// NewShortenHandler returns handlers grouped for router wiring.
func NewShortenHandler(svc *service.ShortenerService) *handlerGroup {
	return &handlerGroup{svc: svc}
}

type handlerGroup struct {
	svc *service.ShortenerService
}

// RegisterRoutes attaches endpoints to chi router
func (h *handlerGroup) RegisterRoutes(r chi.Router) {
	r.Post("/v1/shorten", h.CreateShorten)
	r.Get("/{short}", h.Redirect)
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}

// CreateShorten handles POST /v1/shorten
func (h *handlerGroup) CreateShorten(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	req.URL = strings.TrimSpace(req.URL)
	if req.URL == "" {
		http.Error(w, "url required", http.StatusBadRequest)
		return
	}

	// call service
	short, err := h.svc.Create(ctx, req.URL, req.CustomAlias)
	if err != nil {
		if err == service.ErrAliasTaken {
			http.Error(w, "alias taken", http.StatusConflict)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := shortenResponse{
		ShortURL: strings.TrimRight(h.svc.BaseURL, "/") + "/" + short,
		ShortKey: short,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

// Redirect handles GET /{short}
func (h *handlerGroup) Redirect(w http.ResponseWriter, r *http.Request) {
	short := chi.URLParam(r, "short")
	if strings.TrimSpace(short) == "" {
		http.Error(w, "missing short key", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	long, err := h.svc.Resolve(ctx, short)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, long, http.StatusFound)
}
