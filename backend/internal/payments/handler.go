package payments

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

func (h *Handler) ListByJob(w http.ResponseWriter, r *http.Request) {
	p := identity.MustPrincipal(r.Context())
	if p.IsTechnician() {
		if err := h.svc.selfServeAllowed(r.Context(), p.BusinessID, p.UserID, r.PathValue("id")); err != nil {
			httpapi.WriteAPIError(w, err)
			return
		}
	}
	res, err := h.svc.ListByJob(r.Context(), p.BusinessID, r.PathValue("id"))
	if err != nil {
		httpapi.WriteAPIError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{"data": res})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	p := identity.MustPrincipal(r.Context())
	if p.IsTechnician() {
		if err := h.svc.selfServeAllowed(r.Context(), p.BusinessID, p.UserID, r.PathValue("id")); err != nil {
			httpapi.WriteAPIError(w, err)
			return
		}
	}
	var req CreateRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON body.")
		return
	}
	res, err := h.svc.Create(r.Context(), p.BusinessID, r.PathValue("id"), p.UserID, req)
	if err != nil {
		httpapi.WriteAPIError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, res)
}