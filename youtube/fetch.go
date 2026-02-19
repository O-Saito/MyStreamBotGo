package youtube

import (
	"MyStreamBot/globals"
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
	// A API pode incluir outros níveis; mantenha se precisar (standard, maxres, etc).
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

	/*d := map[string]any{
		"client_id":     globals.GetConfig().YouTubeClientID,
		"client_secret": globals.GetConfig().YouTubeClientSecret,
		"refresh_token": currentuser.RefreshToken,
		"grant_type":    "refresh_token",
	}
	data, _ := json.Marshal(d)
	helpers.Logf(helpers.Red, "YT DATA %v", d)*/
	resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	/*req, _ := http.NewRequest("POST", "https://oauth2.googleapis.com/token", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)*/
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var u OAuthResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &u)
	if u.AccessToken != "" {
		currentuser.Token = u.AccessToken
		currentuser.TokenExpiresIn = u.ExpiresIn
		globals.GetState().SetYouTubeUser(currentuser)
		return nil
	}
	return fmt.Errorf("falha ao executar refresh token! %s: %s", u.Error, u.ErrorDesc)
}

func GetCurrentStreamings() (*LiveBroadcastListResponse, error) {
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/youtube/v3/liveBroadcasts?broadcastStatus=active&part=snippet", nil)
	resp, err := DoYouTubeRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var r LiveBroadcastListResponse
	json.Unmarshal(body, &r)
	return &r, nil
}

func GetCurrentYouTubeChannel() (*YouTubeChannelListResponse, error) {
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/youtube/v3/channels?part=id,snippet&mine=true", nil)
	resp, err := DoYouTubeRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var r YouTubeChannelListResponse
	json.Unmarshal(body, &r)
	return &r, nil
}
