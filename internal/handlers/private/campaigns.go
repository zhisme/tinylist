package private

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/zhisme/tinylist/internal/db"
	"github.com/zhisme/tinylist/internal/handlers/response"
	"github.com/zhisme/tinylist/internal/models"
	"github.com/zhisme/tinylist/internal/worker"
)

// CampaignHandler handles campaign-related requests
type CampaignHandler struct {
	db     *db.DB
	worker *worker.CampaignWorker
}

// NewCampaignHandler creates a new campaign handler
func NewCampaignHandler(database *db.DB, w *worker.CampaignWorker) *CampaignHandler {
	return &CampaignHandler{db: database, worker: w}
}

// CreateCampaignRequest represents the request body for creating a campaign
type CreateCampaignRequest struct {
	Subject  string  `json:"subject"`
	BodyText string  `json:"body_text"`
	BodyHTML *string `json:"body_html,omitempty"`
}

// UpdateCampaignRequest represents the request body for updating a campaign
type UpdateCampaignRequest struct {
	Subject  *string `json:"subject,omitempty"`
	BodyText *string `json:"body_text,omitempty"`
	BodyHTML *string `json:"body_html,omitempty"`
}

// Create handles POST /api/private/campaigns
func (h *CampaignHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid JSON body")
		return
	}

	// Validate subject
	req.Subject = strings.TrimSpace(req.Subject)
	if req.Subject == "" {
		response.BadRequest(w, "subject is required")
		return
	}
	if len(req.Subject) > 500 {
		response.BadRequest(w, "subject must be 500 characters or less")
		return
	}

	// Validate body_text
	req.BodyText = strings.TrimSpace(req.BodyText)
	if req.BodyText == "" {
		response.BadRequest(w, "body_text is required")
		return
	}

	// Trim body_html if present
	if req.BodyHTML != nil {
		trimmed := strings.TrimSpace(*req.BodyHTML)
		if trimmed == "" {
			req.BodyHTML = nil
		} else {
			req.BodyHTML = &trimmed
		}
	}

	campaign := &models.Campaign{
		UUID:     uuid.New().String(),
		Subject:  req.Subject,
		BodyText: req.BodyText,
		BodyHTML: req.BodyHTML,
		Status:   models.CampaignStatusDraft,
	}

	if err := h.db.CreateCampaign(campaign); err != nil {
		response.InternalError(w, "failed to create campaign")
		return
	}

	response.Created(w, campaign)
}

// List handles GET /api/private/campaigns
func (h *CampaignHandler) List(w http.ResponseWriter, r *http.Request) {
	campaigns, err := h.db.ListCampaigns()
	if err != nil {
		response.InternalError(w, "failed to list campaigns")
		return
	}

	// Ensure we return an empty array instead of null
	if campaigns == nil {
		campaigns = []*models.Campaign{}
	}

	response.OK(w, campaigns)
}

// Get handles GET /api/private/campaigns/{id}
func (h *CampaignHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "campaign id is required")
		return
	}

	campaign, err := h.db.GetCampaignByUUID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "failed to get campaign") {
			response.NotFound(w, "campaign not found")
			return
		}
		response.InternalError(w, "failed to get campaign")
		return
	}

	response.OK(w, campaign)
}

// Update handles PUT /api/private/campaigns/{id}
func (h *CampaignHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "campaign id is required")
		return
	}

	// Get existing campaign
	campaign, err := h.db.GetCampaignByUUID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "failed to get campaign") {
			response.NotFound(w, "campaign not found")
			return
		}
		response.InternalError(w, "failed to get campaign")
		return
	}

	// Only allow editing draft campaigns
	if campaign.Status != models.CampaignStatusDraft {
		response.BadRequest(w, "can only edit draft campaigns")
		return
	}

	var req UpdateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid JSON body")
		return
	}

	// Update fields if provided
	if req.Subject != nil {
		subject := strings.TrimSpace(*req.Subject)
		if subject == "" {
			response.BadRequest(w, "subject cannot be empty")
			return
		}
		if len(subject) > 500 {
			response.BadRequest(w, "subject must be 500 characters or less")
			return
		}
		campaign.Subject = subject
	}

	if req.BodyText != nil {
		bodyText := strings.TrimSpace(*req.BodyText)
		if bodyText == "" {
			response.BadRequest(w, "body_text cannot be empty")
			return
		}
		campaign.BodyText = bodyText
	}

	if req.BodyHTML != nil {
		trimmed := strings.TrimSpace(*req.BodyHTML)
		if trimmed == "" {
			campaign.BodyHTML = nil
		} else {
			campaign.BodyHTML = &trimmed
		}
	}

	if err := h.db.UpdateCampaign(campaign); err != nil {
		response.InternalError(w, "failed to update campaign")
		return
	}

	response.OK(w, campaign)
}

// Delete handles DELETE /api/private/campaigns/{id}
func (h *CampaignHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "campaign id is required")
		return
	}

	// Get campaign to find internal ID
	campaign, err := h.db.GetCampaignByUUID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "failed to get campaign") {
			response.NotFound(w, "campaign not found")
			return
		}
		response.InternalError(w, "failed to get campaign")
		return
	}

	// Only allow deleting draft campaigns
	if campaign.Status != models.CampaignStatusDraft {
		response.BadRequest(w, "can only delete draft campaigns")
		return
	}

	if err := h.db.DeleteCampaign(campaign.ID); err != nil {
		response.InternalError(w, "failed to delete campaign")
		return
	}

	response.NoContent(w)
}

// Send handles POST /api/private/campaigns/{id}/send
func (h *CampaignHandler) Send(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "campaign id is required")
		return
	}

	// Get campaign to find internal ID
	campaign, err := h.db.GetCampaignByUUID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "failed to get campaign") {
			response.NotFound(w, "campaign not found")
			return
		}
		response.InternalError(w, "failed to get campaign")
		return
	}

	// Check if already sending
	if h.worker.IsSending(campaign.ID) {
		response.BadRequest(w, "campaign is already being sent")
		return
	}

	// Check campaign status
	if campaign.Status != models.CampaignStatusDraft {
		response.BadRequest(w, "can only send draft campaigns")
		return
	}

	// Start sending in background
	go func() {
		if err := h.worker.SendCampaign(campaign.ID); err != nil {
			log.Printf("Campaign %s send failed: %v", campaign.UUID, err)
		}
	}()

	response.Accepted(w, map[string]string{
		"message": "campaign sending started",
		"id":      campaign.UUID,
	})
}

// Cancel handles POST /api/private/campaigns/{id}/cancel
func (h *CampaignHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "campaign id is required")
		return
	}

	// Get campaign to find internal ID
	campaign, err := h.db.GetCampaignByUUID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "failed to get campaign") {
			response.NotFound(w, "campaign not found")
			return
		}
		response.InternalError(w, "failed to get campaign")
		return
	}

	// Check if campaign is sending
	if !h.worker.IsSending(campaign.ID) {
		response.BadRequest(w, "campaign is not currently sending")
		return
	}

	// Cancel the campaign
	if err := h.worker.CancelCampaign(campaign.ID); err != nil {
		response.InternalError(w, "failed to cancel campaign")
		return
	}

	response.OK(w, map[string]string{
		"message": "campaign cancellation requested",
		"id":      campaign.UUID,
	})
}

// Journal handles GET /api/private/campaigns/{id}/journal
func (h *CampaignHandler) Journal(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "campaign id is required")
		return
	}

	// Get campaign to find internal ID
	campaign, err := h.db.GetCampaignByUUID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "failed to get campaign") {
			response.NotFound(w, "campaign not found")
			return
		}
		response.InternalError(w, "failed to get campaign")
		return
	}

	journal, err := h.db.GetCampaignJournal(campaign.ID)
	if err != nil {
		response.InternalError(w, "failed to get campaign journal")
		return
	}

	// Ensure we return an empty array instead of null
	if journal == nil {
		journal = []*models.CampaignJournal{}
	}

	response.OK(w, journal)
}

// Routes returns a router with all campaign routes
func (h *CampaignHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/send", h.Send)
	r.Post("/{id}/cancel", h.Cancel)
	r.Get("/{id}/journal", h.Journal)
	return r
}
