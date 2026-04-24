package twitch

import (
	"MyStreamBot/helpers"
	
	"time"
)

var urlAPIVideos = "https://api.tv/helix/videos"

type Video struct {
	ID            string              `json:"id"`
	StreamID      string              `json:"stream_id"`
	UserID        string              `json:"user_id"`
	UserLogin     string              `json:"user_login"`
	UserName      string              `json:"user_name"`
	Title         string              `json:"title"`
	Description   string              `json:"description"`
	CreatedAt     time.Time           `json:"created_at"`
	PublishedAt   time.Time           `json:"published_at"`
	URL           string              `json:"url"`
	ThumbnailURL  string              `json:"thumbnail_url"`
	Viewable      string              `json:"viewable"`
	ViewCount     int                 `json:"view_count"`
	Language      string              `json:"language"`
	Type          string              `json:"type"`
	Duration      string              `json:"duration"`
	MutedSegments []VideoMutedSegment `json:"muted_segments"`
}

type VideoMutedSegment struct {
	Duration int `json:"duration"`
	Offset   int `json:"offset"`
}

type GetVideosRequest struct {
	VideoIDs []string
	UserID   string
	GameID   string
	Period   string
	Sort     string
	Type     string
}

type GetVideosResponse struct {
	Data       []Video           `json:"data"`
	Pagination Pagination `json:"pagination"`
}

func GetVideos(req GetVideosRequest) (*PaginationData[Video], error) {
	opts := map[string]any{}

	if req.VideoIDs != nil {
		opts["id"] = req.VideoIDs
	}
	if req.UserID != "" {
		opts["user_id"] = req.UserID
	}
	if req.GameID != "" {
		opts["game_id"] = req.GameID
	}
	if req.Period != "" {
		opts["period"] = req.Period
	}
	if req.Sort != "" {
		opts["sort"] = req.Sort
	}
	if req.Type != "" {
		opts["type"] = req.Type
	}

	url := BuildURL(HelixBaseURL+"/videos", opts)

	result, err := ExecuteRequest[GetVideosResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetVideos: userID=%v, gameID=%v", req.UserID, req.GameID)
		return nil, err
	}

	pagData := &PaginationData[Video]{
		Data: result.Data,
	}
	pagData.Pagination.Cursor = result.Pagination.Cursor

	return pagData, nil
}

func DeleteVideos(videoIDs []string) error {
	opts := map[string]any{}
	for _, id := range videoIDs {
		opts["id"] = id
	}

	url := BuildURL(urlAPIVideos, opts)

	_, err := ExecuteRequest[struct{}]("DELETE", url, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] DeleteVideos: videoIDs=%v", videoIDs)
		return err
	}

	return nil
}
