package models

import "time"

// Subscriber represents an email subscriber
type Subscriber struct {
	ID               int        `json:"-"`
	UUID             string     `json:"id"`
	Email            string     `json:"email"`
	Name             string     `json:"name"`
	Status           string     `json:"status"` // pending, verified, unsubscribed
	VerifyToken      *string    `json:"-"`
	UnsubscribeToken string     `json:"-"`
	CreatedAt        time.Time  `json:"created_at"`
	VerifiedAt       *time.Time `json:"verified_at,omitempty"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// SubscriberStatus constants
const (
	StatusPending      = "pending"
	StatusVerified     = "verified"
	StatusUnsubscribed = "unsubscribed"
)
