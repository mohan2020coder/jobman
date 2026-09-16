package auth

import (
	"encoding/json"
	"net/http"

	"github.com/jobman/backend/pkg/httpapi"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON body.")
		return
	}
	res, err := h.svc.Register(r.Context(), req)
	if err != nil {
		httpapi.WriteAPIError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, res)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON body.")
		return
	}
	res, err := h.svc.Login(r.Context(), req)
	if err != nil {
		httpapi.WriteAPIError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	p := MustPrincipal(r.Context())
	res, err := h.svc.Me(r.Context(), p)
	if err != nil {
		httpapi.WriteAPIError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, res)
}

// Logout is a no-op for stateless JWTs; the client discards the token.
func (h *Handler) Logout(w http.ResponseWriter, _ *http.Request) {
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{"message": "Logged out successfully."})
}