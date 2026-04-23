package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
	"time"
)

var urlAPIStreams = "https://api.twitch.tv/helix/streams"
var urlAPIStreamKey = "https://api.twitch.tv/helix/streams/key"
var urlAPIStreamMarkers = "https://api.twitch.tv/helix/streams/markers"

type Stream struct {
	ID            string    `json:"id"`
	UserID       string    `json:"user_id"`
	UserLogin   string    `json:"user_login"`
	UserName    string    `json:"user_name"`
	GameID      string    `json:"game_id"`
	GameName   string    `json:"game_name"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Tags        []string  `json:"tags"`
	ViewerCount int       `json:"viewer_count"`
	StartedAt   time.Time  `json:"started_at"`
	Language   string    `json:"language"`
	ThumbnailURL string   `json:"thumbnail_url"`
	TagIDs     []string   `json:"tag_ids"`
	IsMature   bool      `json:"is_mature"`
}

type StreamMarker struct {
	ID            string    `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	Description   string    `json:"description"`
	PositionSeconds int     `json:"position_seconds"`
}

type StreamKey struct {
	StreamKey string `json:"stream_key"`
}

type GetStreamsResponse struct {
	Data       []Stream    `json:"data"`
	Pagination Pagination `json:"pagination"`
}

type GetStreamMarkersResponse struct {
	Data []struct {
		VideoID   string         `json:"video_id"`
		CreatedAt time.Time      `json:"created_at"`
		Markers  []StreamMarker `json:"markers"`
	} `json:"data"`
}

func GetStreamKey() (string, error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
	}

	url := twitch.BuildURL(urlAPIStreamKey, opts)

	result, err := twitch.ExecuteRequest[struct {
		Data []StreamKey `json:"data"`
	}]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetStreamKey: broadcasterID=%v", user.UserID)
		return "", err
	}

	if len(result.Data) == 0 {
		return "", nil
	}

	return result.Data[0].StreamKey, nil
}

func GetStreams(userIDs []string, gameIDs []string, languages []string, req *twitch.PaginationRequest) (*twitch.PaginationData[Stream], error) {
	opts := map[string]any{}

	for _, id := range userIDs {
		opts["user_id"] = id
	}
	for _, id := range gameIDs {
		opts["game_id"] = id
	}
	for _, lang := range languages {
		opts["language"] = lang
	}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/streams", opts)

	result, err := twitch.ExecuteRequest[twitch.PaginationData[Stream]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetStreams: userIDs=%v", userIDs)
		return nil, err
	}

	result.GetNext = func() *twitch.PaginationData[Stream] {
		GetStreams(userIDs, gameIDs, languages, &twitch.PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: req.Quantity,
		})
		return result
	}

	return result, nil
}

func GetFollowedStreams(userID string, req *twitch.PaginationRequest) (*twitch.PaginationData[Stream], error) {
	user := globals.GetState().GetTwitchUser()
	if userID == "" {
		userID = user.UserID
	}

	opts := map[string]any{
		"user_id": userID,
	}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/streams/followed", opts)

	result, err := twitch.ExecuteRequest[twitch.PaginationData[Stream]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetFollowedStreams: userID=%v", userID)
		return nil, err
	}

	result.GetNext = func() *twitch.PaginationData[Stream] {
		GetFollowedStreams(userID, &twitch.PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: req.Quantity,
		})
		return result
	}

	return result, nil
}

type CreateStreamMarkerRequest struct {
	Description string `json:"description,omitempty"`
}

func CreateStreamMarker(req CreateStreamMarkerRequest) error {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/streams/markers", opts)

	_, err := twitch.ExecuteJSONRequest[struct{}, CreateStreamMarkerRequest]("POST", url, req, 201)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] CreateStreamMarker: broadcasterID=%v", user.UserID)
		return err
	}

	return nil
}

func GetStreamMarkers(videoID string, req *twitch.PaginationRequest) (*twitch.PaginationData[GetStreamMarkersResponse], error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
		"video_id":       videoID,
	}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/streams/markers", opts)

	result, err := twitch.ExecuteRequest[twitch.PaginationData[GetStreamMarkersResponse]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetStreamMarkers: broadcasterID=%v", user.UserID)
		return nil, err
	}

	result.GetNext = func() *twitch.PaginationData[GetStreamMarkersResponse] {
		GetStreamMarkers(videoID, &twitch.PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: req.Quantity,
		})
		return result
	}

	return result, nil
}