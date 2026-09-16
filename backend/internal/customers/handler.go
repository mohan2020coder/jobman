package customers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jobman/backend/internal/identity"
	"github.com/jobman/backend/pkg/httpapi"
)

type Handler struct {
	svc *Service
	// jobs of a customer are resolved via the jobs service; set to non-nil
	// to enable GET /:id/jobs.
	jobs func(ctx context.Context, businessID, customerID string) (any, error)
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) SetJobsLister(l func(ctx context.Context, businessID, customerID string) (any, error)) { h.jobs = l }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	p := identity.MustPrincipal(r.Context())
	page, limit := httpapi.ParsePageLimit(r, 20)
	res, err := h.svc.List(r.Context(), p.BusinessID, ListFilters{
		Search: r.URL.Query().Get("search"),
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		httpapi.WriteAPIError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{
		"data": res.Data,
		"pagination": httpapi.Pagination{
			Page: page, Limit: limit, Total: res.Total,
		},
	})
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
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{"message": "Customer deleted."})
}

func (h *Handler) Jobs(w http.ResponseWriter, r *http.Request) {
	p := identity.MustPrincipal(r.Context())
	if h.jobs == nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Customer jobs unavailable.")
		return
	}
	res, err := h.jobs(r.Context(), p.BusinessID, r.PathValue("id"))
	if err != nil {
		httpapi.WriteAPIError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{"data": res})
}