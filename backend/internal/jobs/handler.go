package jobs

import (
	"context"
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

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	p := identity.MustPrincipal(r.Context())
	page, limit := httpapi.ParsePageLimit(r, 20)
	q := r.URL.Query()
	res, err := h.svc.List(r.Context(), p.BusinessID, ListFilters{
		Status:       q.Get("status"),
		TechnicianID: q.Get("technician_id"),
		CustomerID:   q.Get("customer_id"),
		Date:         q.Get("date"),
		Search:       q.Get("search"),
		SortBy:       q.Get("sort_by"),
		Page:         page,
		Limit:        limit,
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
	res, err := h.svc.Create(r.Context(), p.BusinessID, p.UserID, req)
	if err != nil {
		httpapi.WriteAPIError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, res)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	p := identity.MustPrincipal(r.Context())
	if p.IsTechnician() {
		if err := h.svc.EnsureAssigned(r.Context(), p.BusinessID, p.UserID, r.PathValue("id")); err != nil {
			httpapi.WriteAPIError(w, err)
			return
		}
	}
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

// Delete soft-cancels a job. Only owner/admin can delete PENDING jobs.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	p := identity.MustPrincipal(r.Context())
	res, err := h.svc.Cancel(r.Context(), p.BusinessID, p.UserID, r.PathValue("id"))
	if err != nil {
		httpapi.WriteAPIError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{"message": "Job cancelled.", "job": res})
}

func (h *Handler) action(next func(ctx context.Context, businessID, userID, jobID string) (*Job, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := identity.MustPrincipal(r.Context())
		res, err := next(r.Context(), p.BusinessID, p.UserID, r.PathValue("id"))
		if err != nil {
			httpapi.WriteAPIError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, res)
	}
}

func (h *Handler) Accept(w http.ResponseWriter, r *http.Request) {
	h.action(h.svc.Accept)(w, r)
}
func (h *Handler) OnTheWay(w http.ResponseWriter, r *http.Request) {
	h.action(h.svc.OnTheWay)(w, r)
}
func (h *Handler) Start(w http.ResponseWriter, r *http.Request) {
	h.action(h.svc.Start)(w, r)
}
func (h *Handler) CancelAction(w http.ResponseWriter, r *http.Request) {
	h.action(h.svc.Cancel)(w, r)
}

func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	p := identity.MustPrincipal(r.Context())
	var req CompleteRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON body.")
		return
	}
	res, err := h.svc.Complete(r.Context(), p.BusinessID, p.UserID, r.PathValue("id"), req)
	if err != nil {
		httpapi.WriteAPIError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, res)
}

// MyJobs serves GET /api/v1/technician/jobs (assigned to the authed tech).
func (h *Handler) MyJobs(w http.ResponseWriter, r *http.Request) {
	p := identity.MustPrincipal(r.Context())
	filter := r.URL.Query().Get("filter")
	switch filter {
	case "today", "upcoming", "completed":
	default:
		filter = ""
	}
	res, err := h.svc.MyJobs(r.Context(), p.BusinessID, p.UserID, filter)
	if err != nil {
		httpapi.WriteAPIError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{"data": res})
}