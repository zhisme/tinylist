package worker

import (
	"context"
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

// campaignContext holds the context and cancel func for a sending campaign
type campaignContext struct {
	cancel context.CancelFunc
}

// CampaignWorker handles sending campaigns
type CampaignWorker struct {
	db        *db.DB
	mailer    *mailer.Mailer
	config    config.SendingConfig
	publicURL string
	mu        sync.Mutex
	sending   map[int]*campaignContext // Track campaigns currently being sent
}

// NewCampaignWorker creates a new campaign worker
func NewCampaignWorker(database *db.DB, mail *mailer.Mailer, cfg config.SendingConfig, publicURL string) *CampaignWorker {
	return &CampaignWorker{
		db:        database,
		mailer:    mail,
		config:    cfg,
		publicURL: publicURL,
		sending:   make(map[int]*campaignContext),
	}
}

// ReplaceTemplateVars replaces {{name}} and {{email}} in text
func ReplaceTemplateVars(text, name, email string) string {
	result := strings.ReplaceAll(text, "{{name}}", name)
	result = strings.ReplaceAll(result, "{{email}}", email)
	return result
}

// logJournal is a helper to log a journal entry
func (w *CampaignWorker) logJournal(campaignID int, eventType, message string) {
	entry := &models.CampaignJournal{
		CampaignID: campaignID,
		EventType:  eventType,
		Message:    message,
	}
	if err := w.db.CreateCampaignJournal(entry); err != nil {
		log.Printf("Warning: failed to create journal entry: %v", err)
	}
}

// SendCampaign starts sending a campaign to all verified subscribers
func (w *CampaignWorker) SendCampaign(campaignID int) error {
	// Check if already sending
	w.mu.Lock()
	if w.sending[campaignID] != nil {
		w.mu.Unlock()
		return fmt.Errorf("campaign %d is already being sent", campaignID)
	}
	ctx, cancel := context.WithCancel(context.Background())
	w.sending[campaignID] = &campaignContext{cancel: cancel}
	w.mu.Unlock()

	defer func() {
		w.mu.Lock()
		delete(w.sending, campaignID)
		w.mu.Unlock()
	}()

	// Get campaign
	campaign, err := w.db.GetCampaignByID(campaignID)
	if err != nil {
		w.logJournal(campaignID, models.JournalEventError, fmt.Sprintf("Failed to get campaign: %v", err))
		return fmt.Errorf("failed to get campaign: %w", err)
	}

	// Check campaign status
	if campaign.Status != models.CampaignStatusDraft {
		w.logJournal(campaignID, models.JournalEventError, "Campaign is not in draft status")
		return fmt.Errorf("campaign is not in draft status")
	}

	// Get all verified subscribers
	subscribers, err := w.db.GetVerifiedSubscribers()
	if err != nil {
		w.logJournal(campaignID, models.JournalEventError, fmt.Sprintf("Failed to get subscribers: %v", err))
		return fmt.Errorf("failed to get subscribers: %w", err)
	}

	if len(subscribers) == 0 {
		w.logJournal(campaignID, models.JournalEventError, "No verified subscribers to send to")
		return fmt.Errorf("no verified subscribers to send to")
	}

	// Log start
	w.logJournal(campaignID, models.JournalEventInfo, fmt.Sprintf("Started sending to %d subscribers", len(subscribers)))

	// Update campaign status to sending
	if err := w.db.UpdateCampaignStatus(campaignID, models.CampaignStatusSending); err != nil {
		w.logJournal(campaignID, models.JournalEventError, fmt.Sprintf("Failed to update status: %v", err))
		return fmt.Errorf("failed to update campaign status: %w", err)
	}

	// Set total count
	if err := w.db.UpdateCampaignCounts(campaignID, len(subscribers), 0, 0); err != nil {
		log.Printf("Warning: failed to update campaign counts: %v", err)
	}

	// Send emails with rate limiting
	sentCount := 0
	failedCount := 0
	cancelled := false
	ticker := time.NewTicker(time.Second / time.Duration(w.config.RateLimit))
	defer ticker.Stop()

	for _, sub := range subscribers {
		// Check for cancellation before each send
		select {
		case <-ctx.Done():
			cancelled = true
			w.logJournal(campaignID, models.JournalEventWarning, fmt.Sprintf("Cancelled: %d sent, %d failed, %d remaining", sentCount, failedCount, len(subscribers)-sentCount-failedCount))
			break
		case <-ticker.C:
			// Continue with rate limiting
		}

		if cancelled {
			break
		}

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
			sendErr = w.mailer.SendCampaign(ctx, sub.Email, sub.Name, subject, bodyText, bodyHTML, unsubscribeURL)
			if sendErr == nil {
				break
			}
			// Check if context was cancelled - don't retry in that case
			if ctx.Err() != nil {
				break
			}
			if attempt < w.config.MaxRetries {
				time.Sleep(w.config.RetryDelay)
			}
		}

		// Check if cancelled during send
		if ctx.Err() != nil {
			cancelled = true
			w.logJournal(campaignID, models.JournalEventWarning, fmt.Sprintf("Cancelled: %d sent, %d failed, %d remaining", sentCount, failedCount, len(subscribers)-sentCount-failedCount))
			break
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
	var finalStatus string
	if cancelled {
		finalStatus = models.CampaignStatusCancelled
	} else if failedCount > 0 && sentCount == 0 {
		finalStatus = models.CampaignStatusFailed
	} else {
		finalStatus = models.CampaignStatusSent
	}
	if err := w.db.UpdateCampaignStatus(campaignID, finalStatus); err != nil {
		w.logJournal(campaignID, models.JournalEventError, fmt.Sprintf("Failed to update final status: %v", err))
		return fmt.Errorf("failed to update final campaign status: %w", err)
	}

	// Log completion
	if cancelled {
		log.Printf("Campaign %d cancelled: %d sent, %d failed", campaignID, sentCount, failedCount)
	} else if failedCount == 0 {
		w.logJournal(campaignID, models.JournalEventSuccess, fmt.Sprintf("Completed: %d emails sent successfully", sentCount))
	} else if sentCount == 0 {
		w.logJournal(campaignID, models.JournalEventError, fmt.Sprintf("Failed: all %d emails failed to send", failedCount))
	} else {
		w.logJournal(campaignID, models.JournalEventWarning, fmt.Sprintf("Completed with errors: %d sent, %d failed", sentCount, failedCount))
	}

	if !cancelled {
		log.Printf("Campaign %d completed: %d sent, %d failed", campaignID, sentCount, failedCount)
	}
	return nil
}

// IsSending returns true if a campaign is currently being sent
func (w *CampaignWorker) IsSending(campaignID int) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.sending[campaignID] != nil
}

// CancelCampaign cancels a currently sending campaign
func (w *CampaignWorker) CancelCampaign(campaignID int) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	ctx := w.sending[campaignID]
	if ctx == nil {
		return fmt.Errorf("campaign %d is not currently sending", campaignID)
	}
	ctx.cancel()
	return nil
}
