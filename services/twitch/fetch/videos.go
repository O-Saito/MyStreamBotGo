package twitch

import (
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
	"time"
)

var urlAPIVideos = "https://api.twitch.tv/helix/videos"

type Video struct {
	ID             string    `json:"id"`
	StreamID      string    `json:"stream_id"`
	BroadcasterID string    `json:"broadcaster_id"`
	BroadcasterLogin string `json:"broadcaster_login"`
	BroadcasterName string `json:"broadcaster_name"`
	Title         string    `json:"title"`
	Description  string    `json:"description"`
	CreatedAt    time.Time  `json:"created_at"`
	PublishedAt  time.Time  `json:"published_at"`
	URL          string    `json:"url"`
	ThumbnailURL string    `json:"thumbnail_url"`
	Viewable     string    `json:"viewable"`
	ViewCount    int       `json:"view_count"`
	Language     string    `json:"language"`
	Type         string    `json:"type"`
	Duration    int       `json:"duration"`
	MutedSegments []VideoMutedSegment `json:"muted_segments"`
}

type VideoMutedSegment struct {
	Duration int `json:"duration"`
	Offset   int `json:"offset"`
}

type GetVideosResponse struct {
	Data       []Video       `json:"data"`
	Pagination Pagination   `json:"pagination"`
}

func GetVideos(videoIDs []string, userID, gameID string, req *twitch.PaginationRequest) (*twitch.PaginationData[Video], error) {
	opts := map[string]any{}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/videos", opts)

	for _, id := range videoIDs {
		url += "&id=" + id
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
			Quantity: req.Quantity,
		})
		return result
	}

	return result, nil
}

func DeleteVideos(videoIDs []string) error {
	opts := map[string]any{}
	for _, id := range videoIDs {
		opts["id"] = id
	}

	url := twitch.BuildURL(urlAPIVideos, opts)

	_, err := twitch.ExecuteRequest[struct{}]("DELETE", url, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] DeleteVideos: videoIDs=%v", videoIDs)
		return err
	}

	return nil
}