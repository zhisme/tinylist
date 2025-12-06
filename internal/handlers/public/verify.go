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

// VerifyHandler handles email verification
type VerifyHandler struct {
	db *db.DB
}

// NewVerifyHandler creates a new verify handler
func NewVerifyHandler(database *db.DB) *VerifyHandler {
	return &VerifyHandler{db: database}
}

// VerifyResponse represents the verification response
type VerifyResponse struct {
	Message string `json:"message"`
}

// Verify handles GET /api/verify/:token
func (h *VerifyHandler) Verify(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		response.BadRequest(w, "verification token is required")
		return
	}

	// Find subscriber by token
	sub, err := h.db.GetSubscriberByVerifyToken(token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "failed to get subscriber") {
			response.NotFound(w, "invalid or expired verification link")
			return
		}
		response.InternalError(w, "verification failed")
		return
	}

	// Check if already verified
	if sub.Status == models.StatusVerified {
		response.OK(w, VerifyResponse{
			Message: "Your email is already verified.",
		})
		return
	}

	// Check if unsubscribed
	if sub.Status == models.StatusUnsubscribed {
		response.BadRequest(w, "this email has been unsubscribed")
		return
	}

	// Update status to verified
	if err := h.db.UpdateSubscriberStatus(sub.ID, models.StatusVerified); err != nil {
		response.InternalError(w, "verification failed")
		return
	}

	response.OK(w, VerifyResponse{
		Message: "Your email has been verified. Thank you for subscribing!",
	})
}
