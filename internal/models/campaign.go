package models

import "time"

// Campaign represents an email campaign
type Campaign struct {
	ID          int        `json:"-"`
	UUID        string     `json:"id"`
	Subject     string     `json:"subject"`
	BodyText    string     `json:"body_text"`
	BodyHTML    *string    `json:"body_html,omitempty"`
	Status      string     `json:"status"` // draft, sending, sent, failed
	TotalCount  int        `json:"total_count"`
	SentCount   int        `json:"sent_count"`
	FailedCount int        `json:"failed_count"`
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// CampaignStatus constants
const (
	CampaignStatusDraft   = "draft"
	CampaignStatusSending = "sending"
	CampaignStatusSent    = "sent"
	CampaignStatusFailed  = "failed"
)

// CampaignLog represents a log entry for campaign sends
type CampaignLog struct {
	ID           int       `json:"id"`
	CampaignID   int       `json:"campaign_id"`
	SubscriberID int       `json:"subscriber_id"`
	Status       string    `json:"status"` // sent, failed
	Error        *string   `json:"error,omitempty"`
	SentAt       time.Time `json:"sent_at"`
}
