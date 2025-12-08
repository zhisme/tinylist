package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/zhisme/tinylist/internal/models"
)

// parseTime parses a SQLite datetime string to time.Time
func parseTime(s string) time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", s)
	return t
}

// parseTimePtr parses a nullable SQLite datetime string to *time.Time
func parseTimePtr(s sql.NullString) *time.Time {
	if !s.Valid {
		return nil
	}
	t := parseTime(s.String)
	return &t
}

// Subscriber queries

// CreateSubscriber inserts a new subscriber
func (db *DB) CreateSubscriber(sub *models.Subscriber) error {
	query := `
		INSERT INTO subscribers (uuid, email, name, status, verify_token, unsubscribe_token, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))
		RETURNING id, created_at, updated_at
	`
	var createdAt, updatedAt string
	err := db.QueryRow(query, sub.UUID, sub.Email, sub.Name, sub.Status, sub.VerifyToken, sub.UnsubscribeToken).Scan(&sub.ID, &createdAt, &updatedAt)
	if err != nil {
		return fmt.Errorf("failed to create subscriber: %w", err)
	}
	sub.CreatedAt = parseTime(createdAt)
	sub.UpdatedAt = parseTime(updatedAt)
	return nil
}

// GetSubscriberByID retrieves a subscriber by ID
func (db *DB) GetSubscriberByID(id int) (*models.Subscriber, error) {
	query := `
		SELECT id, uuid, email, name, status, verify_token, unsubscribe_token,
		       created_at, verified_at, updated_at
		FROM subscribers
		WHERE id = ?
	`
	var sub models.Subscriber
	var createdAt, updatedAt string
	var verifiedAt sql.NullString
	err := db.QueryRow(query, id).Scan(
		&sub.ID, &sub.UUID, &sub.Email, &sub.Name, &sub.Status,
		&sub.VerifyToken, &sub.UnsubscribeToken,
		&createdAt, &verifiedAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriber: %w", err)
	}
	sub.CreatedAt = parseTime(createdAt)
	sub.UpdatedAt = parseTime(updatedAt)
	sub.VerifiedAt = parseTimePtr(verifiedAt)
	return &sub, nil
}

// GetSubscriberByUUID retrieves a subscriber by UUID
func (db *DB) GetSubscriberByUUID(uuid string) (*models.Subscriber, error) {
	query := `
		SELECT id, uuid, email, name, status, verify_token, unsubscribe_token,
		       created_at, verified_at, updated_at
		FROM subscribers
		WHERE uuid = ?
	`
	var sub models.Subscriber
	var createdAt, updatedAt string
	var verifiedAt sql.NullString
	err := db.QueryRow(query, uuid).Scan(
		&sub.ID, &sub.UUID, &sub.Email, &sub.Name, &sub.Status,
		&sub.VerifyToken, &sub.UnsubscribeToken,
		&createdAt, &verifiedAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriber: %w", err)
	}
	sub.CreatedAt = parseTime(createdAt)
	sub.UpdatedAt = parseTime(updatedAt)
	sub.VerifiedAt = parseTimePtr(verifiedAt)
	return &sub, nil
}

// GetSubscriberByEmail retrieves a subscriber by email
func (db *DB) GetSubscriberByEmail(email string) (*models.Subscriber, error) {
	query := `
		SELECT id, uuid, email, name, status, verify_token, unsubscribe_token,
		       created_at, verified_at, updated_at
		FROM subscribers
		WHERE email = ? COLLATE NOCASE
	`
	var sub models.Subscriber
	var createdAt, updatedAt string
	var verifiedAt sql.NullString
	err := db.QueryRow(query, email).Scan(
		&sub.ID, &sub.UUID, &sub.Email, &sub.Name, &sub.Status,
		&sub.VerifyToken, &sub.UnsubscribeToken,
		&createdAt, &verifiedAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriber: %w", err)
	}
	sub.CreatedAt = parseTime(createdAt)
	sub.UpdatedAt = parseTime(updatedAt)
	sub.VerifiedAt = parseTimePtr(verifiedAt)
	return &sub, nil
}

// GetSubscriberByVerifyToken retrieves a subscriber by verification token
func (db *DB) GetSubscriberByVerifyToken(token string) (*models.Subscriber, error) {
	query := `
		SELECT id, uuid, email, name, status, verify_token, unsubscribe_token,
		       created_at, verified_at, updated_at
		FROM subscribers
		WHERE verify_token = ?
	`
	var sub models.Subscriber
	var createdAt, updatedAt string
	var verifiedAt sql.NullString
	err := db.QueryRow(query, token).Scan(
		&sub.ID, &sub.UUID, &sub.Email, &sub.Name, &sub.Status,
		&sub.VerifyToken, &sub.UnsubscribeToken,
		&createdAt, &verifiedAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriber: %w", err)
	}
	sub.CreatedAt = parseTime(createdAt)
	sub.UpdatedAt = parseTime(updatedAt)
	sub.VerifiedAt = parseTimePtr(verifiedAt)
	return &sub, nil
}

// GetSubscriberByUnsubscribeToken retrieves a subscriber by unsubscribe token
func (db *DB) GetSubscriberByUnsubscribeToken(token string) (*models.Subscriber, error) {
	query := `
		SELECT id, uuid, email, name, status, verify_token, unsubscribe_token,
		       created_at, verified_at, updated_at
		FROM subscribers
		WHERE unsubscribe_token = ?
	`
	var sub models.Subscriber
	var createdAt, updatedAt string
	var verifiedAt sql.NullString
	err := db.QueryRow(query, token).Scan(
		&sub.ID, &sub.UUID, &sub.Email, &sub.Name, &sub.Status,
		&sub.VerifyToken, &sub.UnsubscribeToken,
		&createdAt, &verifiedAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriber: %w", err)
	}
	sub.CreatedAt = parseTime(createdAt)
	sub.UpdatedAt = parseTime(updatedAt)
	sub.VerifiedAt = parseTimePtr(verifiedAt)
	return &sub, nil
}

// ListSubscribers retrieves subscribers with pagination and filtering
func (db *DB) ListSubscribers(status string, page, perPage int) ([]*models.Subscriber, int, error) {
	// Build query with optional status filter
	whereClause := ""
	args := []interface{}{}
	if status != "" {
		whereClause = "WHERE status = ?"
		args = append(args, status)
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM subscribers %s", whereClause)
	var total int
	if err := db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count subscribers: %w", err)
	}

	// Get paginated results
	offset := (page - 1) * perPage
	query := fmt.Sprintf(`
		SELECT id, uuid, email, name, status, verify_token, unsubscribe_token,
		       created_at, verified_at, updated_at
		FROM subscribers
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)
	args = append(args, perPage, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list subscribers: %w", err)
	}
	defer rows.Close()

	var subscribers []*models.Subscriber
	for rows.Next() {
		var sub models.Subscriber
		var createdAt, updatedAt string
		var verifiedAt sql.NullString
		if err := rows.Scan(
			&sub.ID, &sub.UUID, &sub.Email, &sub.Name, &sub.Status,
			&sub.VerifyToken, &sub.UnsubscribeToken,
			&createdAt, &verifiedAt, &updatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan subscriber: %w", err)
		}
		sub.CreatedAt = parseTime(createdAt)
		sub.UpdatedAt = parseTime(updatedAt)
		sub.VerifiedAt = parseTimePtr(verifiedAt)
		subscribers = append(subscribers, &sub)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating subscribers: %w", err)
	}

	return subscribers, total, nil
}

// UpdateSubscriberStatus updates subscriber status and verified_at timestamp
func (db *DB) UpdateSubscriberStatus(id int, status string) error {
	query := `
		UPDATE subscribers
		SET status = ?,
		    verified_at = CASE WHEN ? = 'verified' THEN datetime('now') ELSE verified_at END,
		    updated_at = datetime('now')
		WHERE id = ?
	`
	result, err := db.Exec(query, status, status, id)
	if err != nil {
		return fmt.Errorf("failed to update subscriber status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DeleteSubscriber permanently deletes a subscriber
func (db *DB) DeleteSubscriber(id int) error {
	query := "DELETE FROM subscribers WHERE id = ?"
	result, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete subscriber: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetVerifiedSubscribers retrieves all verified subscribers for campaign sending
func (db *DB) GetVerifiedSubscribers() ([]*models.Subscriber, error) {
	query := `
		SELECT id, uuid, email, name, status, verify_token, unsubscribe_token,
		       created_at, verified_at, updated_at
		FROM subscribers
		WHERE status = 'verified'
		ORDER BY created_at ASC
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get verified subscribers: %w", err)
	}
	defer rows.Close()

	var subscribers []*models.Subscriber
	for rows.Next() {
		var sub models.Subscriber
		var createdAt, updatedAt string
		var verifiedAt sql.NullString
		if err := rows.Scan(
			&sub.ID, &sub.UUID, &sub.Email, &sub.Name, &sub.Status,
			&sub.VerifyToken, &sub.UnsubscribeToken,
			&createdAt, &verifiedAt, &updatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan subscriber: %w", err)
		}
		sub.CreatedAt = parseTime(createdAt)
		sub.UpdatedAt = parseTime(updatedAt)
		sub.VerifiedAt = parseTimePtr(verifiedAt)
		subscribers = append(subscribers, &sub)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating subscribers: %w", err)
	}

	return subscribers, nil
}

// Campaign queries

// CreateCampaign inserts a new campaign
func (db *DB) CreateCampaign(campaign *models.Campaign) error {
	query := `
		INSERT INTO campaigns (uuid, subject, body_text, body_html, status, created_at)
		VALUES (?, ?, ?, ?, ?, datetime('now'))
		RETURNING id
	`
	err := db.QueryRow(query, campaign.UUID, campaign.Subject, campaign.BodyText, campaign.BodyHTML, campaign.Status).Scan(&campaign.ID)
	if err != nil {
		return fmt.Errorf("failed to create campaign: %w", err)
	}
	return nil
}

// GetCampaignByID retrieves a campaign by ID
func (db *DB) GetCampaignByID(id int) (*models.Campaign, error) {
	query := `
		SELECT id, uuid, subject, body_text, body_html, status,
		       total_count, sent_count, failed_count,
		       created_at, started_at, completed_at
		FROM campaigns
		WHERE id = ?
	`
	var c models.Campaign
	var createdAt string
	var startedAt, completedAt sql.NullString
	err := db.QueryRow(query, id).Scan(
		&c.ID, &c.UUID, &c.Subject, &c.BodyText, &c.BodyHTML, &c.Status,
		&c.TotalCount, &c.SentCount, &c.FailedCount,
		&createdAt, &startedAt, &completedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign: %w", err)
	}
	c.CreatedAt = parseTime(createdAt)
	c.StartedAt = parseTimePtr(startedAt)
	c.CompletedAt = parseTimePtr(completedAt)
	return &c, nil
}

// GetCampaignByUUID retrieves a campaign by UUID
func (db *DB) GetCampaignByUUID(uuid string) (*models.Campaign, error) {
	query := `
		SELECT id, uuid, subject, body_text, body_html, status,
		       total_count, sent_count, failed_count,
		       created_at, started_at, completed_at
		FROM campaigns
		WHERE uuid = ?
	`
	var c models.Campaign
	var createdAt string
	var startedAt, completedAt sql.NullString
	err := db.QueryRow(query, uuid).Scan(
		&c.ID, &c.UUID, &c.Subject, &c.BodyText, &c.BodyHTML, &c.Status,
		&c.TotalCount, &c.SentCount, &c.FailedCount,
		&createdAt, &startedAt, &completedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign: %w", err)
	}
	c.CreatedAt = parseTime(createdAt)
	c.StartedAt = parseTimePtr(startedAt)
	c.CompletedAt = parseTimePtr(completedAt)
	return &c, nil
}

// ListCampaigns retrieves all campaigns
func (db *DB) ListCampaigns() ([]*models.Campaign, error) {
	query := `
		SELECT id, uuid, subject, body_text, body_html, status,
		       total_count, sent_count, failed_count,
		       created_at, started_at, completed_at
		FROM campaigns
		ORDER BY created_at DESC
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list campaigns: %w", err)
	}
	defer rows.Close()

	var campaigns []*models.Campaign
	for rows.Next() {
		var c models.Campaign
		var createdAt string
		var startedAt, completedAt sql.NullString
		if err := rows.Scan(
			&c.ID, &c.UUID, &c.Subject, &c.BodyText, &c.BodyHTML, &c.Status,
			&c.TotalCount, &c.SentCount, &c.FailedCount,
			&createdAt, &startedAt, &completedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan campaign: %w", err)
		}
		c.CreatedAt = parseTime(createdAt)
		c.StartedAt = parseTimePtr(startedAt)
		c.CompletedAt = parseTimePtr(completedAt)
		campaigns = append(campaigns, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating campaigns: %w", err)
	}

	return campaigns, nil
}

// UpdateCampaignStatus updates campaign status
func (db *DB) UpdateCampaignStatus(id int, status string) error {
	query := `
		UPDATE campaigns
		SET status = ?,
		    started_at = CASE WHEN ? = 'sending' AND started_at IS NULL THEN datetime('now') ELSE started_at END,
		    completed_at = CASE WHEN ? IN ('sent', 'failed') THEN datetime('now') ELSE completed_at END
		WHERE id = ?
	`
	result, err := db.Exec(query, status, status, status, id)
	if err != nil {
		return fmt.Errorf("failed to update campaign status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// UpdateCampaign updates campaign subject, body_text, and body_html
func (db *DB) UpdateCampaign(campaign *models.Campaign) error {
	query := `
		UPDATE campaigns
		SET subject = ?, body_text = ?, body_html = ?
		WHERE id = ?
	`
	result, err := db.Exec(query, campaign.Subject, campaign.BodyText, campaign.BodyHTML, campaign.ID)
	if err != nil {
		return fmt.Errorf("failed to update campaign: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// UpdateCampaignCounts updates campaign counters
func (db *DB) UpdateCampaignCounts(id, totalCount, sentCount, failedCount int) error {
	query := `
		UPDATE campaigns
		SET total_count = ?,
		    sent_count = ?,
		    failed_count = ?
		WHERE id = ?
	`
	result, err := db.Exec(query, totalCount, sentCount, failedCount, id)
	if err != nil {
		return fmt.Errorf("failed to update campaign counts: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DeleteCampaign permanently deletes a campaign
func (db *DB) DeleteCampaign(id int) error {
	query := "DELETE FROM campaigns WHERE id = ?"
	result, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete campaign: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Campaign log queries

// CreateCampaignLog inserts a campaign log entry
func (db *DB) CreateCampaignLog(log *models.CampaignLog) error {
	query := `
		INSERT INTO campaign_logs (campaign_id, subscriber_id, status, error, sent_at)
		VALUES (?, ?, ?, ?, datetime('now'))
		RETURNING id
	`
	err := db.QueryRow(query, log.CampaignID, log.SubscriberID, log.Status, log.Error).Scan(&log.ID)
	if err != nil {
		return fmt.Errorf("failed to create campaign log: %w", err)
	}
	return nil
}

// GetCampaignLogs retrieves all logs for a campaign
func (db *DB) GetCampaignLogs(campaignID int) ([]*models.CampaignLog, error) {
	query := `
		SELECT id, campaign_id, subscriber_id, status, error, sent_at
		FROM campaign_logs
		WHERE campaign_id = ?
		ORDER BY sent_at DESC
	`
	rows, err := db.Query(query, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign logs: %w", err)
	}
	defer rows.Close()

	var logs []*models.CampaignLog
	for rows.Next() {
		var log models.CampaignLog
		var sentAt string
		if err := rows.Scan(&log.ID, &log.CampaignID, &log.SubscriberID, &log.Status, &log.Error, &sentAt); err != nil {
			return nil, fmt.Errorf("failed to scan campaign log: %w", err)
		}
		log.SentAt = parseTime(sentAt)
		logs = append(logs, &log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating campaign logs: %w", err)
	}

	return logs, nil
}

// Settings queries

// GetSetting retrieves a setting value by key
func (db *DB) GetSetting(key string) (string, error) {
	query := "SELECT value FROM settings WHERE key = ?"
	var value string
	err := db.QueryRow(query, key).Scan(&value)
	if err != nil {
		return "", fmt.Errorf("failed to get setting: %w", err)
	}
	return value, nil
}

// SetSetting sets a setting value
func (db *DB) SetSetting(key, value string) error {
	query := `
		INSERT INTO settings (key, value, updated_at)
		VALUES (?, ?, datetime('now'))
		ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = datetime('now')
	`
	_, err := db.Exec(query, key, value, value)
	if err != nil {
		return fmt.Errorf("failed to set setting: %w", err)
	}
	return nil
}

// GetAllSettings retrieves all settings as a map
func (db *DB) GetAllSettings() (map[string]string, error) {
	query := "SELECT key, value FROM settings"
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all settings: %w", err)
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("failed to scan setting: %w", err)
		}
		settings[key] = value
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating settings: %w", err)
	}

	return settings, nil
}
