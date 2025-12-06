package public

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/zhisme/tinylist/internal/db"
	"github.com/zhisme/tinylist/internal/handlers/response"
	"github.com/zhisme/tinylist/internal/models"
)

// UnsubscribeHandler handles unsubscribe requests
type UnsubscribeHandler struct {
	db *db.DB
}

// NewUnsubscribeHandler creates a new unsubscribe handler
func NewUnsubscribeHandler(database *db.DB) *UnsubscribeHandler {
	return &UnsubscribeHandler{db: database}
}

// UnsubscribeResponse represents the unsubscribe response
type UnsubscribeResponse struct {
	Message string `json:"message"`
}

// Unsubscribe handles GET /api/unsubscribe/:token
func (h *UnsubscribeHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		response.BadRequest(w, "unsubscribe token is required")
		return
	}

	// Find subscriber by token
	sub, err := h.db.GetSubscriberByUnsubscribeToken(token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "failed to get subscriber") {
			response.NotFound(w, "invalid unsubscribe link")
			return
		}
		response.InternalError(w, "unsubscribe failed")
		return
	}

	// Check if already unsubscribed
	if sub.Status == models.StatusUnsubscribed {
		response.OK(w, UnsubscribeResponse{
			Message: "You have already been unsubscribed.",
		})
		return
	}

	// Update status to unsubscribed
	if err := h.db.UpdateSubscriberStatus(sub.ID, models.StatusUnsubscribed); err != nil {
		response.InternalError(w, "unsubscribe failed")
		return
	}

	response.OK(w, UnsubscribeResponse{
		Message: "You have been unsubscribed successfully.",
	})
}
