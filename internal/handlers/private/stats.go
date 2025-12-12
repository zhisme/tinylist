package private

import (
	"net/http"

	"github.com/zhisme/tinylist/internal/db"
	"github.com/zhisme/tinylist/internal/handlers/response"
)

// StatsHandler handles stats API requests
type StatsHandler struct {
	db *db.DB
}

// NewStatsHandler creates a new stats handler
func NewStatsHandler(database *db.DB) *StatsHandler {
	return &StatsHandler{
		db: database,
	}
}

// GetStats returns dashboard statistics
func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.db.GetStats()
	if err != nil {
		response.InternalError(w, "Failed to get stats")
		return
	}

	response.OK(w, stats)
}
