package twitch

import (
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var urlAPIVideos = "https://api.twitch.tv/helix/videos"

type Video struct {
	ID             string    `json:"id"`
	StreamID       string    `json:"stream_id"`
	BroadcasterID  string    `json:"broadcaster_id"`
	BroadcasterLogin string  `json:"broadcaster_login"`
	BroadcasterName string   `json:"broadcaster_name"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	CreatedAt      time.Time `json:"created_at"`
	PublishedAt    time.Time `json:"published_at"`
	URL            string    `json:"url"`
	ThumbnailURL   string    `json:"thumbnail_url"`
	Viewable       string    `json:"viewable"`
	ViewCount      int       `json:"view_count"`
	Language       string    `json:"language"`
	Type           string    `json:"type"`
	Duration       int       `json:"duration"`
	MutedSegments  []VideoMutedSegment `json:"muted_segments"`
}

type VideoMutedSegment struct {
	Duration int `json:"duration"`
	Offset   int `json:"offset"`
}

type GetVideosResponse struct {
	Data       []Video     `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

func GetVideos(videoIDs []string, userID, gameID string, first int, after string) ([]Video, error) {
	url := urlAPIVideos + "?"

	for _, id := range videoIDs {
		url += fmt.Sprintf("&id=%s", id)
	}
	if userID != "" {
		url += "&user_id=" + userID
	}
	if gameID != "" {
		url += "&game_id=" + gameID
	}
	if first > 0 {
		url += fmt.Sprintf("&first=%d", first)
	}
	if after != "" {
		url += "&after=" + after
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetVideos http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetVideos: videoIDs=%v", videoIDs)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetVideos io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("GetVideos: failed: %s", body)
	}

	var result GetVideosResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetVideos io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetVideos json.Unmarshal failed: %v", err)
		return nil, err
	}

	return result.Data, nil
}

func DeleteVideos(videoIDs []string) error {
	url := urlAPIVideos + "?"
	for _, id := range videoIDs {
		url += fmt.Sprintf("&id=%s", id)
	}

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] DeleteVideos http.NewRequest failed: %v", err)
		return err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] DeleteVideos: videoIDs=%v", videoIDs)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] DeleteVideos io.ReadAll failed: %v", err)
			return err
		}
		return fmt.Errorf("DeleteVideos: failed: %s", body)
	}

	return nil
}