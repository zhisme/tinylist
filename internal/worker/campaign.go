package worker

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/zhisme/tinylist/internal/config"
	"github.com/zhisme/tinylist/internal/db"
	"github.com/zhisme/tinylist/internal/mailer"
	"github.com/zhisme/tinylist/internal/models"
)

// CampaignWorker handles sending campaigns
type CampaignWorker struct {
	db        *db.DB
	mailer    *mailer.Mailer
	config    config.SendingConfig
	publicURL string
	mu        sync.Mutex
	sending   map[int]bool // Track campaigns currently being sent
}

// NewCampaignWorker creates a new campaign worker
func NewCampaignWorker(database *db.DB, mail *mailer.Mailer, cfg config.SendingConfig, publicURL string) *CampaignWorker {
	return &CampaignWorker{
		db:        database,
		mailer:    mail,
		config:    cfg,
		publicURL: publicURL,
		sending:   make(map[int]bool),
	}
}

// ReplaceTemplateVars replaces {{name}} and {{email}} in text
func ReplaceTemplateVars(text, name, email string) string {
	result := strings.ReplaceAll(text, "{{name}}", name)
	result = strings.ReplaceAll(result, "{{email}}", email)
	return result
}

// SendCampaign starts sending a campaign to all verified subscribers
func (w *CampaignWorker) SendCampaign(campaignID int) error {
	// Check if already sending
	w.mu.Lock()
	if w.sending[campaignID] {
		w.mu.Unlock()
		return fmt.Errorf("campaign %d is already being sent", campaignID)
	}
	w.sending[campaignID] = true
	w.mu.Unlock()

	defer func() {
		w.mu.Lock()
		delete(w.sending, campaignID)
		w.mu.Unlock()
	}()

	// Get campaign
	campaign, err := w.db.GetCampaignByID(campaignID)
	if err != nil {
		return fmt.Errorf("failed to get campaign: %w", err)
	}

	// Check campaign status
	if campaign.Status != models.CampaignStatusDraft {
		return fmt.Errorf("campaign is not in draft status")
	}

	// Get all verified subscribers
	subscribers, err := w.db.GetVerifiedSubscribers()
	if err != nil {
		return fmt.Errorf("failed to get subscribers: %w", err)
	}

	if len(subscribers) == 0 {
		return fmt.Errorf("no verified subscribers to send to")
	}

	// Update campaign status to sending
	if err := w.db.UpdateCampaignStatus(campaignID, models.CampaignStatusSending); err != nil {
		return fmt.Errorf("failed to update campaign status: %w", err)
	}

	// Set total count
	if err := w.db.UpdateCampaignCounts(campaignID, len(subscribers), 0, 0); err != nil {
		log.Printf("Warning: failed to update campaign counts: %v", err)
	}

	// Send emails with rate limiting
	sentCount := 0
	failedCount := 0
	ticker := time.NewTicker(time.Second / time.Duration(w.config.RateLimit))
	defer ticker.Stop()

	for _, sub := range subscribers {
		<-ticker.C // Rate limit

		// Replace template variables
		subject := ReplaceTemplateVars(campaign.Subject, sub.Name, sub.Email)
		bodyText := ReplaceTemplateVars(campaign.BodyText, sub.Name, sub.Email)
		var bodyHTML string
		if campaign.BodyHTML != nil {
			bodyHTML = ReplaceTemplateVars(*campaign.BodyHTML, sub.Name, sub.Email)
		}

		// Build unsubscribe URL
		unsubscribeURL := fmt.Sprintf("%s/api/unsubscribe/%s", w.publicURL, sub.UnsubscribeToken)

		// Attempt to send with retries
		var sendErr error
		for attempt := 0; attempt <= w.config.MaxRetries; attempt++ {
			sendErr = w.mailer.SendCampaign(sub.Email, sub.Name, subject, bodyText, bodyHTML, unsubscribeURL)
			if sendErr == nil {
				break
			}
			if attempt < w.config.MaxRetries {
				time.Sleep(w.config.RetryDelay)
			}
		}

		// Log the result
		logEntry := &models.CampaignLog{
			CampaignID:   campaignID,
			SubscriberID: sub.ID,
		}

		if sendErr != nil {
			logEntry.Status = "failed"
			errStr := sendErr.Error()
			logEntry.Error = &errStr
			failedCount++
		} else {
			logEntry.Status = "sent"
			sentCount++
		}

		if err := w.db.CreateCampaignLog(logEntry); err != nil {
			log.Printf("Warning: failed to create campaign log: %v", err)
		}

		// Update counts periodically (every batch)
		if (sentCount+failedCount)%w.config.BatchSize == 0 {
			if err := w.db.UpdateCampaignCounts(campaignID, len(subscribers), sentCount, failedCount); err != nil {
				log.Printf("Warning: failed to update campaign counts: %v", err)
			}
		}
	}

	// Final count update
	if err := w.db.UpdateCampaignCounts(campaignID, len(subscribers), sentCount, failedCount); err != nil {
		log.Printf("Warning: failed to update final campaign counts: %v", err)
	}

	// Update campaign status
	finalStatus := models.CampaignStatusSent
	if failedCount > 0 && sentCount == 0 {
		finalStatus = models.CampaignStatusFailed
	}
	if err := w.db.UpdateCampaignStatus(campaignID, finalStatus); err != nil {
		return fmt.Errorf("failed to update final campaign status: %w", err)
	}

	log.Printf("Campaign %d completed: %d sent, %d failed", campaignID, sentCount, failedCount)
	return nil
}

// IsSending returns true if a campaign is currently being sent
func (w *CampaignWorker) IsSending(campaignID int) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.sending[campaignID]
}
