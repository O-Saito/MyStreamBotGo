package twitch

import (
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
)

type Game struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	BoxArtURL string `json:"box_art_url"`
}

type GetTopGamesResponse struct {
	Data       []Game           `json:"data"`
	Pagination twitch.Pagination `json:"pagination"`
}

type GetGamesResponse struct {
	Data []Game `json:"data"`
}

func GetTopGames(req *twitch.PaginationRequest) (*twitch.PaginationData[Game], error) {
	opts := map[string]any{}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/games/top", opts)

	result, err := twitch.ExecuteRequest[twitch.PaginationData[Game]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetTopGames: req=%v", req)
		return nil, err
	}

	result.GetNext = func() *twitch.PaginationData[Game] {
		GetTopGames(&twitch.PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: req.Quantity,
		})
		return result
	}

	return result, nil
}

func GetGames(gameIDs []string, names []string) ([]Game, error) {
	opts := map[string]any{}

	for _, id := range gameIDs {
		opts["id"] = id
	}
	for _, name := range names {
		opts["name"] = name
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/games", opts)

	result, err := twitch.ExecuteRequest[GetGamesResponse]("GET", url, 200)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}