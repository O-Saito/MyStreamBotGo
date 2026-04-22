package twitch

import (
	twitch "MyStreamBot/services/twitch"
	"fmt"
)

type Game struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	BoxArtURL string `json:"box_art_url"`
}

type GetTopGamesResponse struct {
	Data       []Game            `json:"data"`
	Pagination twitch.Pagination `json:"pagination"`
}

type GetGamesResponse struct {
	Data []Game `json:"data"`
}

func GetTopGames(first int, after string) ([]Game, error) {
	url := twitch.HelixBaseURL + "/games/top"
	if first > 0 {
		url += fmt.Sprintf("?first=%d", first)
		if after != "" {
			url += "&after=" + after
		}
	} else if after != "" {
		url += "?after=" + after
	}

	result, err := twitch.ExecuteRequest[GetTopGamesResponse]("GET", url, 200)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

func GetGames(gameIDs []string, names []string) ([]Game, error) {
	url := twitch.HelixBaseURL + "/games?"

	for _, id := range gameIDs {
		url += fmt.Sprintf("id=%s&", id)
	}
	for _, name := range names {
		url += fmt.Sprintf("name=%s&", name)
	}

	result, err := twitch.ExecuteRequest[GetGamesResponse]("GET", url, 200)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}