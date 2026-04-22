package twitch

import (
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
	"fmt"
	"time"
)

var (
	urlAPIBitsLeaderboard       = twitch.HelixBaseURL + "/bits/leaderboard"
	urlAPICheermotes            = twitch.HelixBaseURL + "/bits/cheermotes"
	urlAPIExtensionTransactions = twitch.HelixBaseURL + "/extensions/transactions"
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
	DateRange DateRange              `json:"date_range"`
	Total     int                    `json:"total"`
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

// TODO: Check again with https://dev.twitch.tv/docs/api/reference#get-extension-transactions
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

type GetExtensionTransactionsResponse struct {
	Data       []ExtensionTransaction `json:"data"`
	Pagination twitch.Pagination      `json:"pagination"`
}

// TODO: Add period string no required
// TODO: Add userId "user_id" string no required
func GetBitsLeaderboard(count int, startedAt *time.Time) (*GetBitsLeaderboardResponse, error) {
	url := twitch.HelixBaseURL + "/bits/leaderboard"

	query := ""
	if count > 0 {
		query = fmt.Sprintf("first=%d", count)
	}
	if startedAt != nil {
		if query != "" {
			query += "&"
		}
		query += fmt.Sprintf("started_at=%s", startedAt.Format(time.RFC3339))
	}

	if query != "" {
		url = url + "?" + query
	}

	result, err := twitch.ExecuteRequest[GetBitsLeaderboardResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetBitsLeaderboard: count=%v", count)
		return nil, err
	}

	return result, nil
}

func GetCheermotes(broadcasterID string) (*GetCheermotesResponse, error) {
	url := twitch.HelixBaseURL + "/bits/cheermotes"

	if broadcasterID != "" {
		url = fmt.Sprintf("%s?broadcaster_id=%s", url, broadcasterID)
	}

	result, err := twitch.ExecuteRequest[GetCheermotesResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetCheermotes: broadcasterID=%v", broadcasterID)
		return nil, err
	}

	return result, nil
}

func GetExtensionTransactions(extensionID string, first int, after string) (*GetExtensionTransactionsResponse, error) {
	url := twitch.HelixBaseURL + "/extensions/transactions"

	query := fmt.Sprintf("extension_id=%s", extensionID)
	if first > 0 {
		query += fmt.Sprintf("&first=%d", first)
	}
	if after != "" {
		query += fmt.Sprintf("&after=%s", after)
	}

	url = url + "?" + query

	result, err := twitch.ExecuteRequest[GetExtensionTransactionsResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetExtensionTransactions: extensionID=%v", extensionID)
		return nil, err
	}

	return result, nil
}
