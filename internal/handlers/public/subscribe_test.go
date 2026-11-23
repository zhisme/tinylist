package public_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zhisme/tinylist/internal/db"
	"github.com/zhisme/tinylist/internal/handlers/public"
	"github.com/zhisme/tinylist/internal/mailer"
	"github.com/zhisme/tinylist/internal/models"
)

// ----------------------------------------------------------------------------
// Test helpers
// ----------------------------------------------------------------------------

func newTestDB(t *testing.T) *db.DB {
	t.Helper()
	database, err := db.New(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("newTestDB: open: %v", err)
	}
	if err := database.Migrate(); err != nil {
		t.Fatalf("newTestDB: migrate: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}

// newHandler constructs a SubscribeHandler with an unconfigured mailer (no
// emails are sent) and the given requireVerification setting.
func newHandler(database *db.DB, requireVerification bool) *public.SubscribeHandler {
	m := mailer.New() // IsConfigured() == false → no SMTP calls in tests
	return public.NewSubscribeHandler(database, m, "http://localhost:8080", requireVerification)
}

func doPost(t *testing.T, h *public.SubscribeHandler, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/subscribe", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Subscribe(w, req)
	return w
}

type apiResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

func decodeResp(t *testing.T, w *httptest.ResponseRecorder) apiResponse {
	t.Helper()
	var resp apiResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decodeResp: %v (body: %s)", err, w.Body.String())
	}
	return resp
}

// ----------------------------------------------------------------------------
// Input validation
// ----------------------------------------------------------------------------

func TestSubscribe_InvalidJSON_Returns400(t *testing.T) {
	h := newHandler(newTestDB(t), true)
	w := doPost(t, h, "not json at all")
	if w.Code != http.StatusBadRequest {
		t.Errorf("invalid JSON: want 400, got %d", w.Code)
	}
}

func TestSubscribe_InvalidEmail_Returns400(t *testing.T) {
	h := newHandler(newTestDB(t), true)

	cases := []struct {
		name  string
		email string
	}{
		{"empty", ""},
		{"plain_text", "notanemail"},
		{"missing_tld", "user@nodomain"},
		{"at_prefix_only", "@domain.com"},
		{"spaces", "user @example.com"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := `{"email":"` + tc.email + `"}`
			w := doPost(t, h, body)
			if w.Code != http.StatusBadRequest {
				t.Errorf("email %q: want 400, got %d (body: %s)", tc.email, w.Code, w.Body.String())
			}
		})
	}
}

// ----------------------------------------------------------------------------
// require_verification = true  (double opt-in, the default)
// ----------------------------------------------------------------------------

func TestSubscribe_WithVerification_CreatesPendingSubscriber(t *testing.T) {
	database := newTestDB(t)
	h := newHandler(database, true)

	w := doPost(t, h, `{"email":"pending@example.com","name":"Test User"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}

	sub, err := database.GetSubscriberByEmail("pending@example.com")
	if err != nil {
		t.Fatalf("GetSubscriberByEmail: %v", err)
	}
	if sub.Status != models.StatusPending {
		t.Errorf("status: want %q, got %q", models.StatusPending, sub.Status)
	}
}

func TestSubscribe_WithVerification_SetsVerifyToken(t *testing.T) {
	database := newTestDB(t)
	h := newHandler(database, true)

	doPost(t, h, `{"email":"tokentest@example.com"}`)

	sub, err := database.GetSubscriberByEmail("tokentest@example.com")
	if err != nil {
		t.Fatalf("GetSubscriberByEmail: %v", err)
	}
	if sub.VerifyToken == nil || *sub.VerifyToken == "" {
		t.Error("verify_token should be set for pending subscribers")
	}
}

func TestSubscribe_WithVerification_ResponseContainsVerifyPrompt(t *testing.T) {
	h := newHandler(newTestDB(t), true)

	w := doPost(t, h, `{"email":"prompt@example.com"}`)
	resp := decodeResp(t, w)

	if !strings.Contains(strings.ToLower(resp.Message), "verify") {
		t.Errorf("response should mention verification, got: %q", resp.Message)
	}
}

func TestSubscribe_WithVerification_NotVisibleToGetVerifiedSubscribers(t *testing.T) {
	database := newTestDB(t)
	h := newHandler(database, true)

	doPost(t, h, `{"email":"notyet@example.com"}`)

	verified, err := database.GetVerifiedSubscribers()
	if err != nil {
		t.Fatalf("GetVerifiedSubscribers: %v", err)
	}
	for _, sub := range verified {
		if sub.Email == "notyet@example.com" {
			t.Error("pending subscriber must not appear in GetVerifiedSubscribers")
		}
	}
}

// ----------------------------------------------------------------------------
// require_verification = false  (single opt-in)
// ----------------------------------------------------------------------------

func TestSubscribe_WithoutVerification_CreatesVerifiedSubscriber(t *testing.T) {
	database := newTestDB(t)
	h := newHandler(database, false)

	w := doPost(t, h, `{"email":"direct@example.com","name":"Direct User"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}

	sub, err := database.GetSubscriberByEmail("direct@example.com")
	if err != nil {
		t.Fatalf("GetSubscriberByEmail: %v", err)
	}
	if sub.Status != models.StatusVerified {
		t.Errorf("status: want %q, got %q", models.StatusVerified, sub.Status)
	}
}

func TestSubscribe_WithoutVerification_VerifyTokenIsNil(t *testing.T) {
	// No verify token should be generated when verification is skipped — the
	// verify endpoint cannot and should not be used for these subscribers.
	database := newTestDB(t)
	h := newHandler(database, false)

	doPost(t, h, `{"email":"notoken@example.com"}`)

	sub, err := database.GetSubscriberByEmail("notoken@example.com")
	if err != nil {
		t.Fatalf("GetSubscriberByEmail: %v", err)
	}
	if sub.VerifyToken != nil {
		t.Errorf("verify_token should be nil when verification is not required, got %q", *sub.VerifyToken)
	}
}

func TestSubscribe_WithoutVerification_ResponseConfirmsSubscription(t *testing.T) {
	h := newHandler(newTestDB(t), false)

	w := doPost(t, h, `{"email":"confirm@example.com"}`)
	resp := decodeResp(t, w)

	if strings.Contains(strings.ToLower(resp.Message), "verify") {
		t.Errorf("response should not mention verification when not required, got: %q", resp.Message)
	}
	if !strings.Contains(strings.ToLower(resp.Message), "subscribed") {
		t.Errorf("response should confirm subscription, got: %q", resp.Message)
	}
}

func TestSubscribe_WithoutVerification_AppearsInGetVerifiedSubscribers(t *testing.T) {
	// This is the key behaviour: a subscriber created with require_verification=false
	// must be immediately reachable by the campaign worker.
	database := newTestDB(t)
	h := newHandler(database, false)

	doPost(t, h, `{"email":"campaign@example.com"}`)

	verified, err := database.GetVerifiedSubscribers()
	if err != nil {
		t.Fatalf("GetVerifiedSubscribers: %v", err)
	}
	for _, sub := range verified {
		if sub.Email == "campaign@example.com" {
			return // found — test passes
		}
	}
	t.Error("directly-verified subscriber should appear in GetVerifiedSubscribers for campaign delivery")
}

// ----------------------------------------------------------------------------
// Resubscription (previously unsubscribed address)
// ----------------------------------------------------------------------------

func TestSubscribe_Resubscription_WithVerification_GoesToPending(t *testing.T) {
	database := newTestDB(t)
	h := newHandler(database, true)

	doPost(t, h, `{"email":"resub@example.com"}`)
	sub, err := database.GetSubscriberByEmail("resub@example.com")
	if err != nil {
		t.Fatalf("setup: GetSubscriberByEmail: %v", err)
	}
	if err := database.UpdateSubscriberStatus(sub.ID, models.StatusUnsubscribed); err != nil {
		t.Fatalf("setup: UpdateSubscriberStatus: %v", err)
	}

	w := doPost(t, h, `{"email":"resub@example.com"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("resubscribe: want 200, got %d: %s", w.Code, w.Body.String())
	}

	sub, err = database.GetSubscriberByEmail("resub@example.com")
	if err != nil {
		t.Fatalf("GetSubscriberByEmail after resubscribe: %v", err)
	}
	if sub.Status != models.StatusPending {
		t.Errorf("resubscribed status (with verification): want %q, got %q", models.StatusPending, sub.Status)
	}
}

func TestSubscribe_Resubscription_WithoutVerification_GoesToVerified(t *testing.T) {
	database := newTestDB(t)
	h := newHandler(database, false)

	doPost(t, h, `{"email":"resubdirect@example.com"}`)
	sub, err := database.GetSubscriberByEmail("resubdirect@example.com")
	if err != nil {
		t.Fatalf("setup: GetSubscriberByEmail: %v", err)
	}
	if err := database.UpdateSubscriberStatus(sub.ID, models.StatusUnsubscribed); err != nil {
		t.Fatalf("setup: UpdateSubscriberStatus: %v", err)
	}

	w := doPost(t, h, `{"email":"resubdirect@example.com"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("resubscribe: want 200, got %d: %s", w.Code, w.Body.String())
	}

	sub, err = database.GetSubscriberByEmail("resubdirect@example.com")
	if err != nil {
		t.Fatalf("GetSubscriberByEmail after resubscribe: %v", err)
	}
	if sub.Status != models.StatusVerified {
		t.Errorf("resubscribed status (without verification): want %q, got %q", models.StatusVerified, sub.Status)
	}
}

// ----------------------------------------------------------------------------
// Duplicate / existing subscriber (subscriber enumeration protection)
// ----------------------------------------------------------------------------

func TestSubscribe_ExistingPendingSubscriber_Returns200(t *testing.T) {
	// A second subscribe attempt for a pending email must return 200 so that
	// an attacker cannot determine whether an address is already on the list.
	database := newTestDB(t)
	h := newHandler(database, true)

	doPost(t, h, `{"email":"dup@example.com"}`)
	w := doPost(t, h, `{"email":"dup@example.com"}`)

	if w.Code != http.StatusOK {
		t.Errorf("duplicate subscribe: want 200 to prevent enumeration, got %d", w.Code)
	}
}

func TestSubscribe_ExistingVerifiedSubscriber_Returns200(t *testing.T) {
	database := newTestDB(t)
	h := newHandler(database, false) // subscriber goes straight to verified

	doPost(t, h, `{"email":"verified@example.com"}`)
	w := doPost(t, h, `{"email":"verified@example.com"}`)

	if w.Code != http.StatusOK {
		t.Errorf("duplicate subscribe (verified): want 200, got %d", w.Code)
	}
}
