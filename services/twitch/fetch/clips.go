package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
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

func CreateClip(broadcasterID string, hasDelay bool) ([]Clip, error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	opts := map[string]any{
		"broadcaster_id": broadcasterID,
		"has_delay":      hasDelay,
	}
	url := twitch.BuildURL(twitch.HelixBaseURL+"/clips", opts)

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

	opts := map[string]any{
		"broadcaster_id": broadcasterID,
		"video_id":       videoID,
	}
	url := twitch.BuildURL(twitch.HelixBaseURL+"/clips", opts)

	resp, err := twitch.ExecuteRequest[CreateClipResponse]("POST", url, 201)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] CreateClipFromVOD: broadcasterID=%v, videoID=%v", broadcasterID, videoID)
		return nil, err
	}

	return resp.Data, nil
}

func GetClips(broadcasterID, gameID string, clipIDs []string, startedAt, endedAt *time.Time, req *twitch.PaginationRequest) ([]Clip, error) {
	opts := map[string]any{}

	if broadcasterID != "" {
		opts["broadcaster_id"] = broadcasterID
	}
	if gameID != "" {
		opts["game_id"] = gameID
	}
	for _, id := range clipIDs {
		opts["id"] = id
	}
	if startedAt != nil {
		opts["started_at"] = startedAt.Format(time.RFC3339)
	}
	if endedAt != nil {
		opts["ended_at"] = endedAt.Format(time.RFC3339)
	}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/clips", opts)

	resp, err := twitch.ExecuteRequest[twitch.PaginationData[Clip]]("GET", url, 200)
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

func GetClipsDownload(clipIDs []string) ([]ClipDownload, error) {
	opts := map[string]any{}
	for _, id := range clipIDs {
		opts["id"] = id
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/clips/download", opts)

	resp, err := twitch.ExecuteRequest[struct {
		Data []ClipDownload `json:"data"`
	}]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetClipsDownload: clipIDs=%v", clipIDs)
		return nil, err
	}

	return resp.Data, nil
}