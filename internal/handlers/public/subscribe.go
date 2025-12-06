package public

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/zhisme/tinylist/internal/db"
	"github.com/zhisme/tinylist/internal/handlers/response"
	"github.com/zhisme/tinylist/internal/mailer"
	"github.com/zhisme/tinylist/internal/models"
)

// SubscribeHandler handles public subscription requests
type SubscribeHandler struct {
	db        *db.DB
	mailer    *mailer.Mailer
	publicURL string
}

// NewSubscribeHandler creates a new subscribe handler
func NewSubscribeHandler(database *db.DB, m *mailer.Mailer, publicURL string) *SubscribeHandler {
	return &SubscribeHandler{
		db:        database,
		mailer:    m,
		publicURL: strings.TrimSuffix(publicURL, "/"),
	}
}

// SubscribeRequest represents the request body for subscribing
// TODO: verify name field, maybe not needed at all, or later can be enriched by user configuration via UI
type SubscribeRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// SubscribeResponse represents the response for subscribing
type SubscribeResponse struct {
	Message string `json:"message"`
}

// emailRegex validates email format
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// Subscribe handles POST /api/subscribe
func (h *SubscribeHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	var req SubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid JSON body")
		return
	}

	// Validate email
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" {
		response.BadRequest(w, "email is required")
		return
	}
	if len(req.Email) > 254 || !emailRegex.MatchString(req.Email) {
		response.BadRequest(w, "invalid email format")
		return
	}

	// Trim name
	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) > 255 {
		req.Name = req.Name[:255]
	}

	// Check for existing subscriber
	existing, err := h.db.GetSubscriberByEmail(req.Email)
	if err == nil && existing != nil {
		// Already subscribed - return success without revealing if they exist
		response.OK(w, SubscribeResponse{
			Message: "Please check your email to verify your subscription.",
		})
		return
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) && !strings.Contains(err.Error(), "failed to get subscriber") {
		response.InternalError(w, "subscription failed")
		return
	}

	// Generate tokens
	verifyToken := uuid.New().String()
	unsubscribeToken := uuid.New().String()

	// Create subscriber
	sub := &models.Subscriber{
		UUID:             uuid.New().String(),
		Email:            req.Email,
		Name:             req.Name,
		Status:           models.StatusPending,
		VerifyToken:      &verifyToken,
		UnsubscribeToken: unsubscribeToken,
	}

	if err := h.db.CreateSubscriber(sub); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			// Race condition - subscriber created between check and insert
			response.OK(w, SubscribeResponse{
				Message: "Please check your email to verify your subscription.",
			})
			return
		}
		response.InternalError(w, "subscription failed")
		return
	}

	// Send verification email
	if h.mailer.IsConfigured() {
		verifyURL := h.publicURL + "/api/verify/" + verifyToken
		name := req.Name
		if name == "" {
			name = "there"
		}
		if err := h.mailer.SendVerification(req.Email, name, verifyURL); err != nil {
			// Log error but don't fail the request
			// In production, we'd want proper logging here
		}
	}

	response.OK(w, SubscribeResponse{
		Message: "Please check your email to verify your subscription.",
	})
}
