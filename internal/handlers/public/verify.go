package public

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/zhisme/tinylist/internal/db"
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

// renderHTML renders a simple HTML page with the given title, message, and status
func renderHTML(w http.ResponseWriter, statusCode int, title, message string, isSuccess bool) {
	iconColor := "#22c55e" // green
	icon := "✓"
	if !isSuccess {
		iconColor = "#ef4444" // red
		icon = "✕"
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            padding: 20px;
        }
        .card {
            background: white;
            border-radius: 16px;
            padding: 48px;
            text-align: center;
            max-width: 420px;
            box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
        }
        .icon {
            width: 80px;
            height: 80px;
            border-radius: 50%%;
            background: %s;
            color: white;
            font-size: 40px;
            display: flex;
            align-items: center;
            justify-content: center;
            margin: 0 auto 24px;
        }
        h1 {
            color: #1f2937;
            font-size: 24px;
            margin-bottom: 12px;
        }
        p {
            color: #6b7280;
            font-size: 16px;
            line-height: 1.6;
        }
    </style>
</head>
<body>
    <div class="card">
        <div class="icon">%s</div>
        <h1>%s</h1>
        <p>%s</p>
    </div>
</body>
</html>`, title, iconColor, icon, title, message)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write([]byte(html))
}

// Verify handles GET /api/verify/:token
func (h *VerifyHandler) Verify(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		renderHTML(w, http.StatusBadRequest, "Invalid Link", "The verification link is missing a token.", false)
		return
	}

	// Find subscriber by token
	sub, err := h.db.GetSubscriberByVerifyToken(token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "failed to get subscriber") {
			renderHTML(w, http.StatusNotFound, "Invalid Link", "This verification link is invalid or has expired.", false)
			return
		}
		renderHTML(w, http.StatusInternalServerError, "Error", "Something went wrong. Please try again later.", false)
		return
	}

	// Check if already verified
	if sub.Status == models.StatusVerified {
		renderHTML(w, http.StatusOK, "Already Verified", "Your email address has already been verified.", true)
		return
	}

	// Check if unsubscribed
	if sub.Status == models.StatusUnsubscribed {
		renderHTML(w, http.StatusBadRequest, "Unsubscribed", "This email address has been unsubscribed from our list.", false)
		return
	}

	// Update status to verified
	if err := h.db.UpdateSubscriberStatus(sub.ID, models.StatusVerified); err != nil {
		renderHTML(w, http.StatusInternalServerError, "Error", "Something went wrong. Please try again later.", false)
		return
	}

	log.Printf(`{"event":"email_verified","email":"%s","status":"verified"}`, sub.Email)

	renderHTML(w, http.StatusOK, "Email Verified", "Thank you! Your email address has been verified successfully.", true)
}
