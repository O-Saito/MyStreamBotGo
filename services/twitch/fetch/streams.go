package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var urlAPIStreams = "https://api.twitch.tv/helix/streams"
var urlAPIStreamKey = "https://api.twitch.tv/helix/streams/key"
var urlAPIStreamMarkers = "https://api.twitch.tv/helix/streams/markers"

type Stream struct {
	ID                   string   `json:"id"`
	UserID               string   `json:"user_id"`
	UserLogin            string   `json:"user_login"`
	UserName             string   `json:"user_name"`
	GameID               string   `json:"game_id"`
	GameName             string   `json:"game_name"`
	Type                 string   `json:"type"`
	Title                string   `json:"title"`
	Tags                 []string `json:"tags"`
	ViewerCount          int      `json:"viewer_count"`
	StartedAt            time.Time `json:"started_at"`
	Language             string   `json:"language"`
	ThumbnailURL         string   `json:"thumbnail_url"`
	TagIDs               []string `json:"tag_ids"`
	IsMature             bool     `json:"is_mature"`
}

type StreamMarker struct {
	ID          string    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	Description string    `json:"description"`
	PositionSeconds int   `json:"position_seconds"`
}

type StreamKey struct {
	StreamKey string `json:"stream_key"`
}

type GetStreamsResponse struct {
	Data       []Stream    `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

type GetStreamMarkersResponse struct {
	Data []struct {
		VideoID   string        `json:"video_id"`
		CreatedAt time.Time     `json:"created_at"`
		Markers   []StreamMarker `json:"markers"`
	} `json:"data"`
}

func GetStreamKey(broadcasterID string) (string, error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s?broadcaster_id=%s", urlAPIStreamKey, broadcasterID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetStreamKey http.NewRequest failed: %v", err)
		return "", err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetStreamKey: broadcasterID=%v", broadcasterID)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetStreamKey io.ReadAll failed: %v", err)
			return "", err
		}
		return "", fmt.Errorf("GetStreamKey: failed: %s", body)
	}

	var result struct {
		Data []StreamKey `json:"data"`
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetStreamKey io.ReadAll failed: %v", err)
		return "", err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetStreamKey json.Unmarshal failed: %v", err)
		return "", err
	}

	if len(result.Data) == 0 {
		return "", nil
	}

	return result.Data[0].StreamKey, nil
}

func GetStreams(userIDs []string, gameIDs []string, languages []string, first int, after string) ([]Stream, error) {
	url := urlAPIStreams + "?"

	for _, id := range userIDs {
		url += fmt.Sprintf("&user_id=%s", id)
	}
	for _, id := range gameIDs {
		url += fmt.Sprintf("&game_id=%s", id)
	}
	for _, lang := range languages {
		url += fmt.Sprintf("&language=%s", lang)
	}
	if first > 0 {
		url += fmt.Sprintf("&first=%d", first)
	}
	if after != "" {
		url += "&after=" + after
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetStreams http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetStreams: userIDs=%v", userIDs)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetStreams io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("GetStreams: failed: %s", body)
	}

	var result GetStreamsResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetStreams io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetStreams json.Unmarshal failed: %v", err)
		return nil, err
	}

	return result.Data, nil
}

func GetFollowedStreams(userID string, first int, after string) ([]Stream, error) {
	user := globals.GetState().GetTwitchUser()
	if userID == "" {
		userID = user.UserID
	}

	url := fmt.Sprintf("%s/followed?user_id=%s", urlAPIStreams, userID)
	if first > 0 {
		url += fmt.Sprintf("&first=%d", first)
	}
	if after != "" {
		url += "&after=" + after
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetFollowedStreams http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetFollowedStreams: userID=%v", userID)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetFollowedStreams io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("GetFollowedStreams: failed: %s", body)
	}

	var result GetStreamsResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetFollowedStreams io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetFollowedStreams json.Unmarshal failed: %v", err)
		return nil, err
	}

	return result.Data, nil
}

type CreateStreamMarkerRequest struct {
	Description string `json:"description,omitempty"`
}

func CreateStreamMarker(broadcasterID string, req CreateStreamMarkerRequest) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	data, err := json.Marshal(req)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] CreateStreamMarker json.Marshal failed: %v", err)
		return err
	}

	url := fmt.Sprintf("%s?broadcaster_id=%s", urlAPIStreamMarkers, broadcasterID)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] CreateStreamMarker http.NewRequest failed: %v", err)
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := twitch.DoRequest(httpReq)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] CreateStreamMarker: broadcasterID=%v", broadcasterID)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] CreateStreamMarker io.ReadAll failed: %v", err)
			return err
		}
		return fmt.Errorf("CreateStreamMarker: failed: %s", body)
	}

	return nil
}

func GetStreamMarkers(broadcasterID, videoID string, first int, after string) ([]GetStreamMarkersResponse, error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s?broadcaster_id=%s", urlAPIStreamMarkers, broadcasterID)
	if videoID != "" {
		url += "&video_id=" + videoID
	}
	if first > 0 {
		url += fmt.Sprintf("&first=%d", first)
	}
	if after != "" {
		url += "&after=" + after
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetStreamMarkers http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetStreamMarkers: broadcasterID=%v", broadcasterID)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetStreamMarkers io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("GetStreamMarkers: failed: %s", body)
	}

	var result []GetStreamMarkersResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetStreamMarkers io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetStreamMarkers json.Unmarshal failed: %v", err)
		return nil, err
	}

	return result, nil
}