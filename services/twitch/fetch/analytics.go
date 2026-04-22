package twitch

import (
	"fmt"
	twitch "MyStreamBot/services/twitch"
	"time"
)

type ExtensionAnalyticsReport struct {
	ExtensionID string    `json:"extension_id"`
	URL         string    `json:"url"`
	Type        string    `json:"type"`
	DateRange   DateRange `json:"date_range"`
}

type GameAnalyticsReport struct {
	GameID    string    `json:"game_id"`
	URL       string    `json:"url"`
	Type      string    `json:"type"`
	DateRange DateRange `json:"date_range"`
}

type DateRange struct {
	StartedAt string `json:"started_at"`
	EndedAt   string `json:"ended_at"`
}

type Pagination struct {
	Cursor string `json:"cursor"`
}

type GetExtensionAnalyticsResponse struct {
	Data       []ExtensionAnalyticsReport `json:"data"`
	Pagination Pagination                 `json:"pagination"`
}

type GetGameAnalyticsResponse struct {
	Data       []GameAnalyticsReport `json:"data"`
	Pagination Pagination            `json:"pagination"`
}

func GetExtensionAnalytics(extensionID, reportType string, startedAt, endedAt *time.Time, first int, after string) (*GetExtensionAnalyticsResponse, error) {
	url := twitch.HelixBaseURL + "/analytics/extensions"

	query := ""
	if extensionID != "" {
		query = fmt.Sprintf("extension_id=%s", extensionID)
	}
	if reportType != "" {
		if query != "" {
			query += "&"
		}
		query += fmt.Sprintf("type=%s", reportType)
	}
	if startedAt != nil && endedAt != nil {
		if query != "" {
			query += "&"
		}
		query += fmt.Sprintf("started_at=%s&ended_at=%s", startedAt.Format(time.RFC3339), endedAt.Format(time.RFC3339))
	}
	if first > 0 {
		if query != "" {
			query += "&"
		}
		query += fmt.Sprintf("first=%d", first)
	}
	if after != "" {
		if query != "" {
			query += "&"
		}
		query += fmt.Sprintf("after=%s", after)
	}

	if query != "" {
		url = url + "?" + query
	}

	return twitch.ExecuteRequest[GetExtensionAnalyticsResponse]("GET", url, 200)
}

func GetGameAnalytics(gameID, reportType string, startedAt, endedAt *time.Time, first int, after string) (*GetGameAnalyticsResponse, error) {
	url := twitch.HelixBaseURL + "/analytics/games"

	query := ""
	if gameID != "" {
		query = fmt.Sprintf("game_id=%s", gameID)
	}
	if reportType != "" {
		if query != "" {
			query += "&"
		}
		query += fmt.Sprintf("type=%s", reportType)
	}
	if startedAt != nil && endedAt != nil {
		if query != "" {
			query += "&"
		}
		query += fmt.Sprintf("started_at=%s&ended_at=%s", startedAt.Format(time.RFC3339), endedAt.Format(time.RFC3339))
	}
	if first > 0 {
		if query != "" {
			query += "&"
		}
		query += fmt.Sprintf("first=%d", first)
	}
	if after != "" {
		if query != "" {
			query += "&"
		}
		query += fmt.Sprintf("after=%s", after)
	}

	if query != "" {
		url = url + "?" + query
	}

	return twitch.ExecuteRequest[GetGameAnalyticsResponse]("GET", url, 200)
}