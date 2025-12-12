package models

// Stats holds dashboard statistics
type Stats struct {
	TotalSubscribers    int `json:"totalSubscribers"`
	VerifiedSubscribers int `json:"verifiedSubscribers"`
	PendingSubscribers  int `json:"pendingSubscribers"`
	TotalCampaigns      int `json:"totalCampaigns"`
	SentCampaigns       int `json:"sentCampaigns"`
}
