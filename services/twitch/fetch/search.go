package twitch

import (
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
	"fmt"
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

func SearchCategories(query string, req *twitch.PaginationRequest) (*twitch.PaginationData[SearchCategory], error) {
	opts := twitch.RequestOptions{}

	if req != nil {
		opts.After = req.Cursor
		opts.First = req.Quantity
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/search/categories", opts)
	url += fmt.Sprintf("&query=%s", query)

	result, err := twitch.ExecuteRequest[twitch.PaginationData[SearchCategory]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] SearchCategories: query=%v", query)
		return nil, err
	}

	result.GetNext = func() *twitch.PaginationData[SearchCategory] {
		SearchCategories(query, &twitch.PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: opts.First,
		})
		return result
	}

	return result, nil
}

func SearchChannels(query string, liveOnly bool, req *twitch.PaginationRequest) (*twitch.PaginationData[SearchChannel], error) {
	opts := twitch.RequestOptions{}

	if req != nil {
		opts.After = req.Cursor
		opts.First = req.Quantity
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/search/channels", opts)
	url += fmt.Sprintf("&query=%s", query)
	if liveOnly {
		url += "&live_only=true"
	}

	result, err := twitch.ExecuteRequest[twitch.PaginationData[SearchChannel]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] SearchChannels: query=%v", query)
		return nil, err
	}

	result.GetNext = func() *twitch.PaginationData[SearchChannel] {
		SearchChannels(query, liveOnly, &twitch.PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: opts.First,
		})
		return result
	}

	return result, nil
}