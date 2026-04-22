package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
	"fmt"
	"time"
)

type Clip struct {
	ID              string    `json:"id"`
	URL             string    `json:"url"`
	EmbedURL        string    `json:"embed_url"`
	BroadcasterID   string    `json:"broadcaster_id"`
	BroadcasterName string    `json:"broadcaster_name"`
	BroadcasterLogin string   `json:"broadcaster_login"`
	CreatorID       string    `json:"creator_id"`
	CreatorName     string    `json:"creator_name"`
	CreatorLogin    string    `json:"creator_login"`
	VideoID         string    `json:"video_id"`
	GameID          string    `json:"game_id"`
	Language        string    `json:"language"`
	Title           string    `json:"title"`
	ViewCount       int       `json:"view_count"`
	CreatedAt       time.Time `json:"created_at"`
	Duration        float64   `json:"duration"`
	ThumbnailURL    string    `json:"thumbnail_url"`
}

type CreateClipResponse struct {
	Data []Clip `json:"data"`
}

type GetClipsResponse struct {
	Data       []Clip           `json:"data"`
	Pagination twitch.Pagination `json:"pagination"`
}

func CreateClip(broadcasterID string, hasDelay bool) ([]Clip, error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s/clips?broadcaster_id=%s", twitch.HelixBaseURL, broadcasterID)
	if hasDelay {
		url += "&has_delay=true"
	}

	resp, err := twitch.ExecuteRequest[CreateClipResponse]("POST", url, 201)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] CreateClip: broadcasterID=%v", broadcasterID)
		return nil, err
	}

	return resp.Data, nil
}

func CreateClipFromVOD(broadcasterID, videoID string) ([]Clip, error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s/clips?broadcaster_id=%s&video_id=%s", twitch.HelixBaseURL, broadcasterID, videoID)

	resp, err := twitch.ExecuteRequest[CreateClipResponse]("POST", url, 201)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] CreateClipFromVOD: broadcasterID=%v, videoID=%v", broadcasterID, videoID)
		return nil, err
	}

	return resp.Data, nil
}

func GetClips(broadcasterID, gameID string, clipIDs []string, startedAt, endedAt *time.Time, first int, after string) ([]Clip, error) {
	url := twitch.HelixBaseURL + "/clips?"

	if broadcasterID != "" {
		url += fmt.Sprintf("broadcaster_id=%s&", broadcasterID)
	}
	if gameID != "" {
		url += fmt.Sprintf("game_id=%s&", gameID)
	}
	for _, id := range clipIDs {
		url += fmt.Sprintf("id=%s&", id)
	}
	if startedAt != nil {
		url += fmt.Sprintf("started_at=%s&", startedAt.Format(time.RFC3339))
	}
	if endedAt != nil {
		url += fmt.Sprintf("ended_at=%s&", endedAt.Format(time.RFC3339))
	}
	if first > 0 {
		url += fmt.Sprintf("first=%d&", first)
	}
	if after != "" {
		url += "after=" + after
	}

	resp, err := twitch.ExecuteRequest[GetClipsResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetClips: broadcasterID=%v", broadcasterID)
		return nil, err
	}

	return resp.Data, nil
}

type ClipDownload struct {
	ID  string `json:"id"`
	URL string `json:"url"`
	Type string `json:"type"`
}

type GetClipsDownloadResponse struct {
	Data []ClipDownload `json:"data"`
}

func GetClipsDownload(clipIDs []string) ([]ClipDownload, error) {
	url := twitch.HelixBaseURL + "/clips/download?"
	for _, id := range clipIDs {
		url += fmt.Sprintf("id=%s&", id)
	}

	resp, err := twitch.ExecuteRequest[GetClipsDownloadResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetClipsDownload: clipIDs=%v", clipIDs)
		return nil, err
	}

	return resp.Data, nil
}