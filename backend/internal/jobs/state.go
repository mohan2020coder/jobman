package jobs

import (
	"github.com/jobman/backend/pkg/httpapi"
)

// allowedTransitions defines the valid status state machine.
var allowedTransitions = map[string]map[string]bool{
	StatusPending: {
		StatusAccepted:  true, // technician accepts
		StatusCancelled: true, // owner/admin cancels
	},
	StatusAccepted: {
		StatusOnTheWay: true,
		StatusCancelled: true,
	},
	StatusOnTheWay: {
		StatusStarted:  true,
		StatusCancelled: true,
	},
	StatusStarted: {
		StatusCompleted: true,
	},
	StatusCompleted: {},
	StatusCancelled: {},
}

// ValidTransition reports whether moving oldStatus -> newStatus is allowed.
func ValidTransition(oldStatus, newStatus string) bool {
	if m, ok := allowedTransitions[oldStatus]; ok {
		return m[newStatus]
	}
	return false
}

// ValidateTransition returns an API error when a transition is not allowed.
func ValidateTransition(oldStatus, newStatus string) error {
	if !ValidTransition(oldStatus, newStatus) {
		return httpapi.NewAPIError(422, "JOB_INVALID_STATUS",
			"Job cannot move from "+oldStatus+" to "+newStatus+".")
	}
	return nil
}

// transitionTimestamp returns the column to stamp for each status.
func transitionTimestamp(status string) string {
	switch status {
	case StatusAccepted:
		return "accepted_at"
	case StatusOnTheWay:
		return "on_the_way_at"
	case StatusStarted:
		return "started_at"
	case StatusCompleted:
		return "completed_at"
	default:
		return ""
	}
}