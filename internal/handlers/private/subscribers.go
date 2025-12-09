package private

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/zhisme/tinylist/internal/db"
	"github.com/zhisme/tinylist/internal/handlers/response"
	"github.com/zhisme/tinylist/internal/mailer"
	"github.com/zhisme/tinylist/internal/models"
)

// SubscriberHandler handles subscriber-related requests
type SubscriberHandler struct {
	db        *db.DB
	mailer    *mailer.Mailer
	publicURL string
}

// NewSubscriberHandler creates a new subscriber handler
func NewSubscriberHandler(database *db.DB, m *mailer.Mailer, publicURL string) *SubscriberHandler {
	return &SubscriberHandler{
		db:        database,
		mailer:    m,
		publicURL: strings.TrimSuffix(publicURL, "/"),
	}
}

// CreateRequest represents the request body for creating a subscriber
type CreateRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// emailRegex validates email format
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// validateEmail checks if an email is valid
func validateEmail(email string) bool {
	if len(email) > 254 {
		return false
	}
	return emailRegex.MatchString(email)
}

// Create handles POST /api/private/subscribers
func (h *SubscriberHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
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
	if !validateEmail(req.Email) {
		response.BadRequest(w, "invalid email format")
		return
	}

  // TODO: check if name needed at all
	// Trim and validate name
	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) > 255 {
		response.BadRequest(w, "name must be 255 characters or less")
		return
	}

	// Check for existing subscriber
	existing, err := h.db.GetSubscriberByEmail(req.Email)
	if err == nil && existing != nil {
		response.Conflict(w, "subscriber with this email already exists")
		return
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) && !strings.Contains(err.Error(), "failed to get subscriber") {
		response.InternalError(w, "failed to check existing subscriber")
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

  // TODO: check whether we need to return message subscriber already exist, maybe just return 201 always
	if err := h.db.CreateSubscriber(sub); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
      // TODO: check whether we need to return message subscriber already exist, maybe just return 201 always
			response.Conflict(w, "subscriber with this email already exists")
			return
		}
		response.InternalError(w, "failed to create subscriber")
		return
	}

	response.Created(w, sub)
}

// List handles GET /api/private/subscribers
func (h *SubscriberHandler) List(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	status := r.URL.Query().Get("status")
	if status != "" && status != models.StatusPending && status != models.StatusVerified && status != models.StatusUnsubscribed {
		response.BadRequest(w, "invalid status: must be pending, verified, or unsubscribed")
		return
	}

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	perPage := 20
	if pp := r.URL.Query().Get("per_page"); pp != "" {
		if parsed, err := strconv.Atoi(pp); err == nil && parsed > 0 && parsed <= 100 {
			perPage = parsed
		}
	}

	subscribers, total, err := h.db.ListSubscribers(status, page, perPage)
	if err != nil {
		response.InternalError(w, "failed to list subscribers")
		return
	}

	// Ensure we return an empty array instead of null
	if subscribers == nil {
		subscribers = []*models.Subscriber{}
	}

	response.PaginatedResponse(w, subscribers, page, perPage, total)
}

// Get handles GET /api/private/subscribers/{id}
func (h *SubscriberHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "subscriber id is required")
		return
	}

	sub, err := h.db.GetSubscriberByUUID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "failed to get subscriber") {
			response.NotFound(w, "subscriber not found")
			return
		}
		response.InternalError(w, "failed to get subscriber")
		return
	}

	response.OK(w, sub)
}

// Delete handles DELETE /api/private/subscribers/{id}
func (h *SubscriberHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "subscriber id is required")
		return
	}

	// First get the subscriber to find internal ID
	sub, err := h.db.GetSubscriberByUUID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "failed to get subscriber") {
			response.NotFound(w, "subscriber not found")
			return
		}
		response.InternalError(w, "failed to get subscriber")
		return
	}

	if err := h.db.DeleteSubscriber(sub.ID); err != nil {
		response.InternalError(w, "failed to delete subscriber")
		return
	}

	response.NoContent(w)
}

// SendVerification handles POST /api/private/subscribers/{id}/send-verification
func (h *SubscriberHandler) SendVerification(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "subscriber id is required")
		return
	}

	sub, err := h.db.GetSubscriberByUUID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "failed to get subscriber") {
			response.NotFound(w, "subscriber not found")
			return
		}
		response.InternalError(w, "failed to get subscriber")
		return
	}

	// Only allow sending verification for pending subscribers
	if sub.Status != models.StatusPending {
		response.BadRequest(w, "can only send verification email to pending subscribers")
		return
	}

	// Check if mailer is configured
	if !h.mailer.IsConfigured() {
		response.BadRequest(w, "SMTP is not configured")
		return
	}

	// Check if verify token exists
	if sub.VerifyToken == nil || *sub.VerifyToken == "" {
		response.InternalError(w, "subscriber has no verification token")
		return
	}

	verifyURL := h.publicURL + "/api/verify/" + *sub.VerifyToken
	name := sub.Name
	if name == "" {
		name = "there"
	}

	if err := h.mailer.SendVerification(sub.Email, name, verifyURL); err != nil {
		response.InternalError(w, "failed to send verification email")
		return
	}

	response.OK(w, map[string]string{"message": "verification email sent"})
}

// Routes returns a router with all subscriber routes
func (h *SubscriberHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.Get)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/send-verification", h.SendVerification)
	return r
}
