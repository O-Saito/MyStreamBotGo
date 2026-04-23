package twitch

import (
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
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

func GetVideos(videoIDs []string, userID, gameID string, req *twitch.PaginationRequest) (*twitch.PaginationData[Video], error) {
	opts := twitch.RequestOptions{}

	if req != nil {
		opts.After = req.Cursor
		opts.First = req.Quantity
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/videos", opts)

	for _, id := range videoIDs {
		url += fmt.Sprintf("&id=%s", id)
	}
	if userID != "" {
		url += "&user_id=" + userID
	}
	if gameID != "" {
		url += "&game_id=" + gameID
	}

	result, err := twitch.ExecuteRequest[twitch.PaginationData[Video]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetVideos: videoIDs=%v", videoIDs)
		return nil, err
	}

	result.GetNext = func() *twitch.PaginationData[Video] {
		GetVideos(videoIDs, userID, gameID, &twitch.PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: opts.First,
		})
		return result
	}

	return result, nil
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