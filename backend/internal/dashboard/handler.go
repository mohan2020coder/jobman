package dashboard

import (
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

func (h *Handler) Today(w http.ResponseWriter, r *http.Request) {
	p := identity.MustPrincipal(r.Context())
	res, err := h.svc.Today(r.Context(), p.BusinessID)
	if err != nil {
		httpapi.WriteAPIError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, res)
}