package twitch

import (
	twitch "MyStreamBot/services/twitch"
	"time"
)

type ExtensionAnalyticsReport struct {
	ExtensionID string           `json:"extension_id"`
	URL         string           `json:"url"`
	Type        string           `json:"type"`
	DateRange   twitch.DateRange `json:"date_range"`
}

type GameAnalyticsReport struct {
	GameID    string           `json:"game_id"`
	URL       string           `json:"url"`
	Type      string           `json:"type"`
	DateRange twitch.DateRange `json:"date_range"`
}

func GetExtensionAnalytics(extensionID, reportType string, startedAt, endedAt *time.Time, req *twitch.PaginationRequest) ([]ExtensionAnalyticsReport, error) {
	opts := map[string]any{}

	if extensionID != "" {
		opts["extension_id"] = extensionID
	}
	if reportType != "" {
		opts["type"] = reportType
	}
	if startedAt != nil && endedAt != nil {
		opts["started_at"] = startedAt.Format(time.RFC3339)
		opts["ended_at"] = endedAt.Format(time.RFC3339)
	}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/analytics/extensions", opts)

	result, err := twitch.ExecuteRequest[twitch.PaginationData[ExtensionAnalyticsReport]]("GET", url, 200)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

func GetGameAnalytics(gameID, reportType string, startedAt, endedAt *time.Time, req *twitch.PaginationRequest) ([]GameAnalyticsReport, error) {
	opts := map[string]any{}

	if gameID != "" {
		opts["game_id"] = gameID
	}
	if reportType != "" {
		opts["type"] = reportType
	}
	if startedAt != nil && endedAt != nil {
		opts["started_at"] = startedAt.Format(time.RFC3339)
		opts["ended_at"] = endedAt.Format(time.RFC3339)
	}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/analytics/games", opts)

	result, err := twitch.ExecuteRequest[twitch.PaginationData[GameAnalyticsReport]]("GET", url, 200)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}