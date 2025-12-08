package private

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/zhisme/tinylist/internal/db"
	"github.com/zhisme/tinylist/internal/handlers/response"
	"github.com/zhisme/tinylist/internal/mailer"
)

// SMTPSettings represents SMTP configuration in the database
type SMTPSettings struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"password"` // Only returned as "***" if set, never actual value
	FromEmail string `json:"from_email"`
	FromName  string `json:"from_name"`
	TLS       bool   `json:"tls"`
}

// SettingsHandler handles settings API requests
type SettingsHandler struct {
	db     *db.DB
	mailer *mailer.Mailer
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(database *db.DB, m *mailer.Mailer) *SettingsHandler {
	return &SettingsHandler{
		db:     database,
		mailer: m,
	}
}

// Routes returns the settings routes
func (h *SettingsHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/smtp", h.GetSMTPSettings)
	r.Put("/smtp", h.UpdateSMTPSettings)
	r.Post("/smtp/test", h.TestSMTPSettings)

	return r
}

// GetSMTPSettings returns current SMTP settings
func (h *SettingsHandler) GetSMTPSettings(w http.ResponseWriter, r *http.Request) {
	dbSettings, err := h.db.GetAllSettings()
	if err != nil {
		// Return empty settings with defaults on error (new install)
		dbSettings = make(map[string]string)
	}

	smtp := SMTPSettings{
		Host:      dbSettings["smtp_host"],
		Username:  dbSettings["smtp_username"],
		FromEmail: dbSettings["smtp_from_email"],
		FromName:  dbSettings["smtp_from_name"],
		Port:      587, // Default
		TLS:       true, // Default to TLS enabled
	}

	// Parse port if present
	if portStr := dbSettings["smtp_port"]; portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			smtp.Port = port
		}
	}

	// Parse TLS if present
	if tlsStr := dbSettings["smtp_tls"]; tlsStr != "" {
		smtp.TLS = tlsStr == "true"
	}

	// Mask password - only indicate if set
	if dbSettings["smtp_password"] != "" {
		smtp.Password = "***"
	}

	response.JSON(w, http.StatusOK, smtp)
}

// UpdateSMTPSettings updates SMTP settings
func (h *SettingsHandler) UpdateSMTPSettings(w http.ResponseWriter, r *http.Request) {
	var req SMTPSettings
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Host == "" {
		response.BadRequest(w, "SMTP host is required")
		return
	}
	if req.FromEmail == "" {
		response.BadRequest(w, "From email is required")
		return
	}

	// Save settings
	if err := h.db.SetSetting("smtp_host", req.Host); err != nil {
		response.InternalError(w, "Failed to save settings")
		return
	}
	if err := h.db.SetSetting("smtp_port", strconv.Itoa(req.Port)); err != nil {
		response.InternalError(w, "Failed to save settings")
		return
	}
	if err := h.db.SetSetting("smtp_username", req.Username); err != nil {
		response.InternalError(w, "Failed to save settings")
		return
	}
	// Only update password if not masked
	if req.Password != "" && req.Password != "***" {
		if err := h.db.SetSetting("smtp_password", req.Password); err != nil {
			response.InternalError(w, "Failed to save settings")
			return
		}
	}
	if err := h.db.SetSetting("smtp_from_email", req.FromEmail); err != nil {
		response.InternalError(w, "Failed to save settings")
		return
	}
	if err := h.db.SetSetting("smtp_from_name", req.FromName); err != nil {
		response.InternalError(w, "Failed to save settings")
		return
	}
	tlsValue := "false"
	if req.TLS {
		tlsValue = "true"
	}
	if err := h.db.SetSetting("smtp_tls", tlsValue); err != nil {
		response.InternalError(w, "Failed to save settings")
		return
	}

	// Reconfigure mailer with new settings
	h.mailer.Reconfigure(req.Host, req.Port, req.Username, h.getPassword(req.Password), req.FromEmail, req.FromName, req.TLS)

	response.JSON(w, http.StatusOK, map[string]string{"message": "Settings saved successfully"})
}

// getPassword returns the actual password to use
// If the request password is masked, fetch from DB
func (h *SettingsHandler) getPassword(reqPassword string) string {
	if reqPassword != "" && reqPassword != "***" {
		return reqPassword
	}
	// Fetch existing password from DB
	pwd, _ := h.db.GetSetting("smtp_password")
	return pwd
}

// TestSMTPSettings sends a test email
func (h *SettingsHandler) TestSMTPSettings(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if req.Email == "" {
		response.BadRequest(w, "Email address is required")
		return
	}

	if !h.mailer.IsConfigured() {
		response.BadRequest(w, "SMTP is not configured")
		return
	}

	if err := h.mailer.SendTest(req.Email); err != nil {
		response.InternalError(w, "Failed to send test email: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Test email sent successfully"})
}
