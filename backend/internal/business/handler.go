package business

import (
	"encoding/json"
	"net/http"

	"github.com/jobman/backend/internal/identity"
	"github.com/jobman/backend/pkg/httpapi"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	p := identity.MustPrincipal(r.Context())
	res, err := h.svc.Get(r.Context(), p.BusinessID)
	if err != nil {
		httpapi.WriteAPIError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	p := identity.MustPrincipal(r.Context())
	var req UpdateRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON body.")
		return
	}
	res, err := h.svc.Update(r.Context(), p.BusinessID, req)
	if err != nil {
		httpapi.WriteAPIError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, res)
}