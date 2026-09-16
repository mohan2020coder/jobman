package technicians

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jobman/backend/internal/identity"
	"github.com/jobman/backend/pkg/httpapi"
)

type Handler struct {
	svc  *Service
	jobs func(ctx context.Context, businessID, technicianID string) (any, error)
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) SetJobsLister(l func(ctx context.Context, businessID, technicianID string) (any, error)) { h.jobs = l }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	p := identity.MustPrincipal(r.Context())
	res, err := h.svc.List(r.Context(), p.BusinessID)
	if err != nil {
		httpapi.WriteAPIError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{"data": res})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	p := identity.MustPrincipal(r.Context())
	var req CreateRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON body.")
		return
	}
	res, err := h.svc.Create(r.Context(), p.BusinessID, req)
	if err != nil {
		httpapi.WriteAPIError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, res)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	p := identity.MustPrincipal(r.Context())
	res, err := h.svc.Get(r.Context(), p.BusinessID, r.PathValue("id"))
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
	res, err := h.svc.Update(r.Context(), p.BusinessID, r.PathValue("id"), req)
	if err != nil {
		httpapi.WriteAPIError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	p := identity.MustPrincipal(r.Context())
	if err := h.svc.Delete(r.Context(), p.BusinessID, r.PathValue("id")); err != nil {
		httpapi.WriteAPIError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{"message": "Technician deactivated."})
}

func (h *Handler) Jobs(w http.ResponseWriter, r *http.Request) {
	p := identity.MustPrincipal(r.Context())
	if h.jobs == nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Technician jobs unavailable.")
		return
	}
	res, err := h.jobs(r.Context(), p.BusinessID, r.PathValue("id"))
	if err != nil {
		httpapi.WriteAPIError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{"data": res})
}