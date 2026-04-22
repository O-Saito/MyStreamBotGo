package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var urlAPITags = "https://api.twitch.tv/helix/tags"

type StreamTag struct {
	TagID          string `json:"tag_id"`
	IsAuto         bool   `json:"is_auto"`
	LocalizationNames map[string]string `json:"localization_names"`
	LocalizationDescriptions map[string]string `json:"localization_descriptions"`
}

type GetAllStreamTagsResponse struct {
	Data       []StreamTag `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

func GetAllStreamTags(first int, after string) ([]StreamTag, error) {
	url := urlAPITags + "/streams"
	if first > 0 {
		url += fmt.Sprintf("?first=%d", first)
		if after != "" {
			url += "&after=" + after
		}
	} else if after != "" {
		url += "?after=" + after
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetAllStreamTags http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetAllStreamTags: no params")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetAllStreamTags io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("GetAllStreamTags: failed: %s", body)
	}

	var result GetAllStreamTagsResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetAllStreamTags io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetAllStreamTags json.Unmarshal failed: %v", err)
		return nil, err
	}

	return result.Data, nil
}

func GetStreamTags(broadcasterID string) ([]StreamTag, error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s/streams?broadcaster_id=%s", urlAPITags, broadcasterID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetStreamTags http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetStreamTags: broadcasterID=%v", broadcasterID)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetStreamTags io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("GetStreamTags: failed: %s", body)
	}

	var result GetAllStreamTagsResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetStreamTags io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetStreamTags json.Unmarshal failed: %v", err)
		return nil, err
	}

	return result.Data, nil
}