package twitch

import (
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
)

type SearchCategory struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	BoxArtURL string `json:"box_art_url"`
}

type SearchChannel struct {
	BroadcasterLanguage string   `json:"broadcaster_language"`
	BroadcasterLogin    string   `json:"broadcaster_login"`
	DisplayName         string   `json:"display_name"`
	GameID              string   `json:"game_id"`
	GameName            string   `json:"game_name"`
	ID                  string   `json:"id"`
	Live                bool     `json:"is_live"`
	TagIDs              []string `json:"tag_ids"`
	Tags                []string `json:"tags"`
	ThumbnailURL        string   `json:"thumbnail_url"`
	Title               string   `json:"title"`
	StartedAt           string   `json:"started_at"`
}

type GetSearchCategoriesResponse struct {
	Data []SearchCategory `json:"data"`
}

func SearchCategories(query string, req *twitch.PaginationRequest) (*twitch.PaginationData[SearchCategory], error) {
	opts := map[string]any{
		"query": query,
	}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/search/categories", opts)

	result, err := twitch.ExecuteRequest[twitch.PaginationData[SearchCategory]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] SearchCategories: query=%v", query)
		return nil, err
	}

	quantity := 0
	if req != nil {
		quantity = req.Quantity
	}
	result.GetNext = func() *twitch.PaginationData[SearchCategory] {
		r, _ := SearchCategories(query, &twitch.PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: quantity,
		})
		return r
	}

	return result, nil
}

func SearchChannels(query string, liveOnly bool, req *twitch.PaginationRequest) (*twitch.PaginationData[SearchChannel], error) {
	opts := map[string]any{
		"query":      query,
		"live_only": liveOnly,
	}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/search/channels", opts)

	result, err := twitch.ExecuteRequest[twitch.PaginationData[SearchChannel]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] SearchChannels: query=%v", query)
		return nil, err
	}

	quantity := 0
	if req != nil {
		quantity = req.Quantity
	}
	result.GetNext = func() *twitch.PaginationData[SearchChannel] {
		r, _ := SearchChannels(query, liveOnly, &twitch.PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: quantity,
		})
		return r
	}

	return result, nil
}
