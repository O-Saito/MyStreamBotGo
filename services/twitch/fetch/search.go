package twitch

import (
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var urlAPISearch = "https://api.twitch.tv/helix/search"

type SearchCategory struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	BoxArtURL string `json:"box_art_url"`
}

type SearchChannel struct {
	ID              string `json:"id"`
	BroadcasterLogin string `json:"broadcaster_login"`
	BroadcasterName string `json:"broadcaster_name"`
	BroadcasterLanguage string `json:"broadcaster_language"`
	GameID          string `json:"game_id"`
	GameName        string `json:"game_name"`
	Live            bool   `json:"is_live"`
	Tags            []string `json:"tags"`
	ThumbnailURL    string `json:"thumbnail_url"`
	Title           string `json:"title"`
	StartedAt        string `json:"started_at"`
}

type GetSearchCategoriesResponse struct {
	Data []SearchCategory `json:"data"`
}

type GetSearchChannelsResponse struct {
	Data       []SearchChannel `json:"data"`
	Pagination Pagination      `json:"pagination"`
}

func SearchCategories(query string, first int, after string) ([]SearchCategory, error) {
	url := fmt.Sprintf("%s/categories?query=%s", urlAPISearch, query)
	if first > 0 {
		url += fmt.Sprintf("&first=%d", first)
	}
	if after != "" {
		url += "&after=" + after
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] SearchCategories http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] SearchCategories: query=%v", query)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] SearchCategories io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("SearchCategories: failed: %s", body)
	}

	var result GetSearchCategoriesResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] SearchCategories io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] SearchCategories json.Unmarshal failed: %v", err)
		return nil, err
	}

	return result.Data, nil
}

func SearchChannels(query string, first int, after string, liveOnly bool) ([]SearchChannel, error) {
	url := fmt.Sprintf("%s/channels?query=%s", urlAPISearch, query)
	if first > 0 {
		url += fmt.Sprintf("&first=%d", first)
	}
	if after != "" {
		url += "&after=" + after
	}
	if liveOnly {
		url += "&live_only=true"
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] SearchChannels http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] SearchChannels: query=%v", query)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] SearchChannels io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("SearchChannels: failed: %s", body)
	}

	var result GetSearchChannelsResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] SearchChannels io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] SearchChannels json.Unmarshal failed: %v", err)
		return nil, err
	}

	return result.Data, nil
}