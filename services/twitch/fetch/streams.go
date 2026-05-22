package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"

	"time"
)

var urlAPIStreams = "https://api.twitch.tv/helix/streams"
var urlAPIStreamKey = "https://api.twitch.tv/helix/streams/key"
var urlAPIStreamMarkers = "https://api.twitch.tv/helix/streams/markers"

type Stream struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	UserLogin    string    `json:"user_login"`
	UserName     string    `json:"user_name"`
	GameID       string    `json:"game_id"`
	GameName     string    `json:"game_name"`
	Type         string    `json:"type"`
	Title        string    `json:"title"`
	Tags         []string  `json:"tags"`
	ViewerCount  int       `json:"viewer_count"`
	StartedAt    time.Time `json:"started_at"`
	Language     string    `json:"language"`
	ThumbnailURL string    `json:"thumbnail_url"`
	TagIDs       []string  `json:"tag_ids"`
	IsMature     bool      `json:"is_mature"`
}

type StreamMarker struct {
	ID              string    `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	Description     string    `json:"description"`
	PositionSeconds int       `json:"position_seconds"`
}

type StreamKey struct {
	StreamKey string `json:"stream_key"`
}

type GetStreamMarkersResponse struct {
	Data []struct {
		VideoID   string         `json:"video_id"`
		CreatedAt time.Time      `json:"created_at"`
		Markers   []StreamMarker `json:"markers"`
	} `json:"data"`
}

func GetStreamKey() (string, error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
	}

	url := BuildURL(urlAPIStreamKey, opts)

	result, err := ExecuteRequest[struct {
		Data []StreamKey `json:"data"`
	}]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetStreamKey: broadcasterID=%v error=%v", user.UserID, err)
		return "", err
	}

	if len(result.Data) == 0 {
		return "", nil
	}

	return result.Data[0].StreamKey, nil
}

func GetStreams(userIDs, userLogins, gameIDs, languages []string, req *PaginationRequest) (*PaginationData[Stream], error) {
	opts := map[string]any{}

	if userIDs != nil {
		opts["user_id"] = userIDs
	}
	if userLogins != nil {
		opts["user_login"] = userLogins
	}
	if gameIDs != nil {
		opts["game_id"] = gameIDs
	}
	if languages != nil {
		opts["language"] = languages
	}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := BuildURL(HelixBaseURL+"/streams", opts)

	result, err := ExecuteRequest[PaginationData[Stream]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetStreams: userIDs=%v error=%v", userIDs, err)
		return nil, err
	}

	quantity := 0
	if req != nil {
		quantity = req.Quantity
	}
	result.GetNext = func() *PaginationData[Stream] {
		r, _ := GetStreams(userIDs, userLogins, gameIDs, languages, &PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: quantity,
		})
		return r
	}

	return result, nil
}

func GetFollowedStreams(req *PaginationRequest) (*PaginationData[Stream], error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"user_id": user.UserID,
	}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := BuildURL(HelixBaseURL+"/streams/followed", opts)

	result, err := ExecuteRequest[PaginationData[Stream]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetFollowedStreams: userID=%v error=%v", user.UserID, err)
		return nil, err
	}

	quantity := 0
	if req != nil {
		quantity = req.Quantity
	}
	result.GetNext = func() *PaginationData[Stream] {
		r, _ := GetFollowedStreams(&PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: quantity,
		})
		return r
	}

	return result, nil
}

type CreateStreamMarkerResponse struct {
	Data []StreamMarker `json:"data"`
}

func CreateStreamMarker(description string) (*StreamMarker, error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"user_id": user.UserID,
	}

	if description != "" {
		opts["description"] = description
	}

	url := BuildURL(HelixBaseURL+"/streams/markers", opts)

	body := map[string]any{
		"description": description,
	}

	result, err := ExecuteJSONRequest[CreateStreamMarkerResponse, map[string]any]("POST", url, body, 201)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] CreateStreamMarker: broadcasterID=%v error=%v", user.UserID, err)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func GetStreamMarkers(videoID string, req *PaginationRequest) (*PaginationData[GetStreamMarkersResponse], error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"user_id":  user.UserID,
		"video_id": videoID,
	}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := BuildURL(HelixBaseURL+"/streams/markers", opts)

	result, err := ExecuteRequest[PaginationData[GetStreamMarkersResponse]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetStreamMarkers: broadcasterID=%v error=%v", user.UserID, err)
		return nil, err
	}

	result.GetNext = func() *PaginationData[GetStreamMarkersResponse] {
		n, _ := GetStreamMarkers(videoID, &PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: req.Quantity,
		})
		return n
	}

	return result, nil
}
