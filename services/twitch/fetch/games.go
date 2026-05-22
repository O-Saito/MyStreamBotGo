package twitch

import (
	"MyStreamBot/helpers"
	
)

type Game struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	BoxArtURL string `json:"box_art_url"`
}

type GetTopGamesResponse struct {
	Data       []Game           `json:"data"`
	Pagination Pagination `json:"pagination"`
}

type GetGamesResponse struct {
	Data []Game `json:"data"`
}

func GetTopGames(req *PaginationRequest) (*PaginationData[Game], error) {
	opts := map[string]any{}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := BuildURL(HelixBaseURL+"/games/top", opts)

	result, err := ExecuteRequest[PaginationData[Game]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetTopGames: req=%v", req)
		return nil, err
	}

	quantity := 0
	if req != nil {
		quantity = req.Quantity
	}
	result.GetNext = func() *PaginationData[Game] {
		r, _ := GetTopGames(&PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: quantity,
		})
		return r
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

	url := BuildURL(HelixBaseURL+"/games", opts)

	result, err := ExecuteRequest[GetGamesResponse]("GET", url, 200)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}