package twitch

// This is not tested neither validated, I do not see using it

import (
	"MyStreamBot/helpers"
	
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var urlAPITeams = "https://api.tv/helix/teams"

type Team struct {
	ID                 string `json:"id"`
	TeamName           string `json:"team_name"`
	TeamDisplayName    string `json:"team_display_name"`
	BroadcasterID      string `json:"broadcaster_id"`
	BroadcasterName    string `json:"broadcaster_name"`
	BroadcasterLogin   string `json:"broadcaster_login"`
	BackgroundImageURL string `json:"background_image_url"`
	Banner             string `json:"banner"`
	ThumbnailURL       string `json:"thumbnail_url"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
	Info               string `json:"info"`
}

type GetChannelTeamsResponse struct {
	Data []Team `json:"data"`
}

type GetTeamsResponse struct {
	Data []Team `json:"data"`
}

func GetChannelTeams(broadcasterID string) ([]Team, error) {
	url := fmt.Sprintf("%s?broadcaster_id=%s", urlAPITeams, broadcasterID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetChannelTeams http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetChannelTeams: broadcasterID=%v", broadcasterID)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetChannelTeams io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("GetChannelTeams: failed: %s", body)
	}

	var result GetChannelTeamsResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetChannelTeams io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetChannelTeams json.Unmarshal failed: %v", err)
		return nil, err
	}

	return result.Data, nil
}

func GetTeams(teamID string) ([]Team, error) {
	url := fmt.Sprintf("%s?id=%s", urlAPITeams, teamID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetTeams http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetTeams: teamID=%v", teamID)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetTeams io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("GetTeams: failed: %s", body)
	}

	var result GetTeamsResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetTeams io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetTeams json.Unmarshal failed: %v", err)
		return nil, err
	}

	return result.Data, nil
}
