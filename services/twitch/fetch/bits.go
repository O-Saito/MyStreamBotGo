package twitch

import (
	"MyStreamBot/helpers"
	
	"time"
)

type BitsLeaderboardEntry struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
	Rank      int    `json:"rank"`
	Score     int    `json:"score"`
}

type GetBitsLeaderboardResponse struct {
	Data      []BitsLeaderboardEntry `json:"data"`
	DateRange DateRange       `json:"date_range"`
	Total     int                    `json:"total"`
}

func GetBitsLeaderboard(count int, startedAt *time.Time) (*GetBitsLeaderboardResponse, error) {
	opts := map[string]any{}

	if count > 0 {
		opts["first"] = count
	}
	if startedAt != nil {
		opts["started_at"] = startedAt.Format(time.RFC3339)
	}

	url := BuildURL(HelixBaseURL+"/bits/leaderboard", opts)

	result, err := ExecuteRequest[GetBitsLeaderboardResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetBitsLeaderboard: count=%v", count)
		return nil, err
	}

	return result, nil
}

type CheermoteTier struct {
	MinBits        int               `json:"min_bits"`
	ID             string            `json:"id"`
	Color          string            `json:"color"`
	Images         map[string]string `json:"images"`
	CanCheer       bool              `json:"can_cheer"`
	ShowInBitsCard bool              `json:"show_in_bits_card"`
}

type CheermoteGroup struct {
	Prefix       string          `json:"prefix"`
	Tiers        []CheermoteTier `json:"tiers"`
	Type         string          `json:"type"`
	Order        int             `json:"order"`
	LastUpdated  string          `json:"last_updated"`
	IsCharitable bool            `json:"is_charitable"`
}

type GetCheermotesResponse struct {
	Data []CheermoteGroup `json:"data"`
}

func GetCheermotes(broadcasterID string) (*GetCheermotesResponse, error) {
	opts := map[string]any{}

	if broadcasterID != "" {
		opts["broadcaster_id"] = broadcasterID
	}

	url := BuildURL(HelixBaseURL+"/bits/cheermotes", opts)

	result, err := ExecuteRequest[GetCheermotesResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetCheermotes: broadcasterID=%v", broadcasterID)
		return nil, err
	}

	return result, nil
}

type ExtensionTransaction struct {
	ID               string  `json:"id"`
	ExtensionID      string  `json:"extension_id"`
	ExtensionVersion string  `json:"extension_version"`
	Product          Product `json:"product"`
	Currency         string  `json:"currency"`
	Price            int     `json:"price"`
	Time             string  `json:"time"`
}

type Product struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Bits          int    `json:"bits"`
	Sku           string `json:"sku"`
	InDevelopment bool   `json:"in_development"`
}

func GetExtensionTransactions(extensionID string, req *PaginationRequest) ([]ExtensionTransaction, error) {
	opts := map[string]any{
		"extension_id": extensionID,
	}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := BuildURL(HelixBaseURL+"/extensions/transactions", opts)

	result, err := ExecuteRequest[PaginationData[ExtensionTransaction]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetExtensionTransactions: extensionID=%v", extensionID)
		return nil, err
	}

	return result.Data, nil
}