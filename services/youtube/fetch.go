package youtube

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type YouTubeChannelListResponse struct {
	Kind  string           `json:"kind"`
	Etag  string           `json:"etag"`
	Items []YouTubeChannel `json:"items"`
}

type YouTubeChannel struct {
	Kind    string                `json:"kind"`
	Etag    string                `json:"etag"`
	ID      string                `json:"id"`
	Snippet YouTubeChannelSnippet `json:"snippet,omitempty"`
}

type YouTubeChannelSnippet struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	CustomURL   string      `json:"customUrl,omitempty"`
	PublishedAt string      `json:"publishedAt"`
	Thumbnails  *Thumbnails `json:"thumbnails,omitempty"`
	Country     string      `json:"country,omitempty"`
}

type LiveBroadcastListResponse struct {
	Kind  string          `json:"kind"`
	Etag  string          `json:"etag"`
	Items []LiveBroadcast `json:"items"`
}

type LiveBroadcast struct {
	Kind    string                `json:"kind"`
	Etag    string                `json:"etag"`
	ID      string                `json:"id"`
	Snippet *LiveBroadcastSnippet `json:"snippet,omitempty"`
	Status  *LiveBroadcastStatus  `json:"status,omitempty"`
}

type LiveBroadcastStatus struct {
	LifeCycleStatus         string `json:"lifeCycleStatus,omitempty"`
	PrivacyStatus           string `json:"privacyStatus,omitempty"`
	RecordingStatus         string `json:"recordingStatus,omitempty"`
	MadeForKids             bool   `json:"madeForKids,omitempty"`
	SelfDeclaredMadeForKids bool   `json:"selfDeclaredMadeForKids,omitempty"`
}

type LiveBroadcastSnippet struct {
	PublishedAt        *time.Time `json:"publishedAt"`
	ChannelID          string     `json:"channelId"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	ScheduledStartTime *time.Time `json:"scheduledStartTime,omitempty"`
	ActualStartTime    *time.Time `json:"actualStartTime,omitempty"`
	Thumbnails         Thumbnails `json:"thumbnails,omitempty"`
	LiveChatID         string     `json:"liveChatId,omitempty"`
}

type Thumbnails struct {
	Default *Thumbnail `json:"default,omitempty"`
	Medium  *Thumbnail `json:"medium,omitempty"`
	High    *Thumbnail `json:"high,omitempty"`
	// API may include other levels; keep if needed (standard, maxres, etc).
}

type Thumbnail struct {
	URL    string `json:"url"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

func RefreshToken() error {
	currentuser := globals.GetState().GetYouTubeUser()

	data := url.Values{}
	data.Set("client_id", globals.GetConfig().YouTubeClientID)
	data.Set("client_secret", globals.GetConfig().YouTubeClientSecret)
	data.Set("refresh_token", currentuser.RefreshToken)
	data.Set("grant_type", "refresh_token")

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var u OAuthResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[YOUTUBE] RefreshToken: no params")
		helpers.Logf(helpers.ERROR, "[YOUTUBE] RefreshToken io.ReadAll failed: %v", err)
		return err
	}
	if err := json.Unmarshal(body, &u); err != nil {
		helpers.Logf(helpers.DEBUG, "[YOUTUBE] RefreshToken: no params")
		helpers.Logf(helpers.ERROR, "[YOUTUBE] RefreshToken json.Unmarshal failed: %v", err)
		return err
	}
	if u.AccessToken != "" {
		currentuser.Token = u.AccessToken
		currentuser.TokenExpiresIn = u.ExpiresIn
		globals.GetState().SetYouTubeUser(currentuser)
		return nil
	}
	return fmt.Errorf("RefreshToken: failed to execute refresh token: error=%s, description=%s", u.Error, u.ErrorDesc)
}

func GetCurrentStreamings() (*LiveBroadcastListResponse, error) {
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/youtube/v3/liveBroadcasts?broadcastStatus=active&part=snippet,status", nil)
	resp, err := DoYouTubeRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[YOUTUBE] GetCurrentStreamings: no params")
		helpers.Logf(helpers.ERROR, "[YOUTUBE] GetCurrentStreamings io.ReadAll failed: %v", err)
		return nil, err
	}
	helpers.Logf(helpers.DEBUG, "[YT] GetCurrentStreamings: %s", body)
	var r LiveBroadcastListResponse
	if err := json.Unmarshal(body, &r); err != nil {
		helpers.Logf(helpers.DEBUG, "[YOUTUBE] GetCurrentStreamings: no params")
		helpers.Logf(helpers.ERROR, "[YOUTUBE] GetCurrentStreamings json.Unmarshal failed: %v", err)
		return nil, err
	}
	return &r, nil
}

func GetNextStreamings() (*LiveBroadcastListResponse, error) {
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/youtube/v3/liveBroadcasts?broadcastStatus=upcoming&part=snippet,status", nil)
	resp, err := DoYouTubeRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[YOUTUBE] GetNextStreamings: no params")
		helpers.Logf(helpers.ERROR, "[YOUTUBE] GetNextStreamings io.ReadAll failed: %v", err)
		return nil, err
	}
	helpers.Logf(helpers.DEBUG, "[YT] GetCurrentStreamings: %s", body)
	var r LiveBroadcastListResponse
	if err := json.Unmarshal(body, &r); err != nil {
		helpers.Logf(helpers.DEBUG, "[YOUTUBE] GetNextStreamings: no params")
		helpers.Logf(helpers.ERROR, "[YOUTUBE] GetNextStreamings json.Unmarshal failed: %v", err)
		return nil, err
	}
	return &r, nil
}

func GetCurrentYouTubeChannel() (*YouTubeChannelListResponse, error) {
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/youtube/v3/channels?part=id,snippet&mine=true", nil)
	resp, err := DoYouTubeRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[YOUTUBE] GetCurrentYouTubeChannel: no params")
		helpers.Logf(helpers.ERROR, "[YOUTUBE] GetCurrentYouTubeChannel io.ReadAll failed: %v", err)
		return nil, err
	}
	var r YouTubeChannelListResponse
	if err := json.Unmarshal(body, &r); err != nil {
		helpers.Logf(helpers.DEBUG, "[YOUTUBE] GetCurrentYouTubeChannel: no params")
		helpers.Logf(helpers.ERROR, "[YOUTUBE] GetCurrentYouTubeChannel json.Unmarshal failed: %v", err)
		return nil, err
	}
	return &r, nil
}

type VideoListResponse struct {
	Kind  string  `json:"kind"`
	Etag  string  `json:"etag"`
	Items []Video `json:"items"`
}

type Video struct {
	Kind                 string            `json:"kind"`
	ID                   string            `json:"id"`
	LiveStreamingDetails *VideoLiveDetails `json:"liveStreamingDetails,omitempty"`
	Snippet              *VideoSnippet     `json:"snippet,omitempty"`
}

type VideoSnippet struct {
	Title string `json:"title,omitempty"`
}

type VideoLiveDetails struct {
	ConcurrentViewers  string `json:"concurrentViewers,omitempty"`
	ActiveLiveChatID   string `json:"activeLiveChatId,omitempty"`
	ActualStartTime    string `json:"actualStartTime,omitempty"`
	ActualEndTime      string `json:"actualEndTime,omitempty"`
	ScheduledStartTime string `json:"scheduledStartTime,omitempty"`
	ScheduledEndTime   string `json:"scheduledEndTime,omitempty"`
}

type StreamData struct {
	VideoID            string `json:"video_id"`
	Title              string `json:"title"`
	ConcurrentViewers  string `json:"concurrent_viewers"`
	LiveChatID         string `json:"live_chat_id"`
	ActualStartTime    string `json:"actual_start_time"`
	ActualEndTime      string `json:"actual_end_time"`
	ScheduledStartTime string `json:"scheduled_start_time"`
	ScheduledEndTime   string `json:"scheduled_end_time"`
}

func GetStreamData(videoID string) (*StreamData, error) {
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/youtube/v3/videos", nil)
	req.URL.RawQuery = fmt.Sprintf("part=liveStreamingDetails,snippet&id=%s", videoID)

	resp, err := DoYouTubeRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var videoResp VideoListResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[YOUTUBE] GetStreamData: videoID=%v", videoID)
		helpers.Logf(helpers.ERROR, "[YOUTUBE] GetStreamData io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &videoResp); err != nil {
		helpers.Logf(helpers.DEBUG, "[YOUTUBE] GetStreamData: videoID=%v", videoID)
		helpers.Logf(helpers.ERROR, "[YOUTUBE] GetStreamData json.Unmarshal failed: %v", err)
		return nil, err
	}

	if len(videoResp.Items) == 0 {
		return nil, nil
	}

	video := videoResp.Items[0]

	return &StreamData{
		VideoID:            videoID,
		Title:              video.Snippet.Title,
		ConcurrentViewers:  video.LiveStreamingDetails.ConcurrentViewers,
		LiveChatID:         video.LiveStreamingDetails.ActiveLiveChatID,
		ActualStartTime:    video.LiveStreamingDetails.ActualStartTime,
		ActualEndTime:      video.LiveStreamingDetails.ActualEndTime,
		ScheduledStartTime: video.LiveStreamingDetails.ScheduledStartTime,
		ScheduledEndTime:   video.LiveStreamingDetails.ScheduledEndTime,
	}, nil
}
