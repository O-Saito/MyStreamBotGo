package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
)

var (
	urlAPIChannels         = HelixBaseURL + "/channels"
	urlAPIChannelEditors   = HelixBaseURL + "/channels/editors"
	urlAPIFollowedChannels = HelixBaseURL + "/channels/followed"
)

type ChannelEditor struct {
	UserID    string `json:"user_id"`
	UserName  string `json:"user_name"`
	CreatedAt string `json:"created_at"`
}

type ChannelInfo struct {
	BroadcasterID               string   `json:"broadcaster_id"`
	BroadcasterLogin            string   `json:"broadcaster_login"`
	BroadcasterName             string   `json:"broadcaster_name"`
	BroadcasterLanguage         string   `json:"broadcaster_language"`
	GameID                      string   `json:"game_id"`
	GameName                    string   `json:"game_name"`
	Title                       string   `json:"title"`
	Delay                       int      `json:"delay"`
	IsBrandedContent            bool     `json:"is_branded_content"`
	Tags                        []string `json:"tags"`
	ContentClassificationLabels []string `json:"content_classification_labels"`
}

type GetChannelEditorsResponse struct {
	Data []ChannelEditor `json:"data"`
}

type UpdateChannelLabel struct {
	ID        string `json:"id"`
	IsEnabled bool   `json:"is_enabled"`
}

type UpdateChannelRequest struct {
	BroadcasterLanguage         string               `json:"broadcaster_language,omitempty"`
	GameID                      *string              `json:"game_id,omitempty"`
	Title                       *string              `json:"title,omitempty"`
	Delay                       *int                 `json:"delay,omitempty"`
	Tags                        []string             `json:"tags,omitempty"`
	ContentClassificationLabels []UpdateChannelLabel `json:"content_classification_labels,omitempty"`
	IsBrandedContent            bool                 `json:"is_branded_content,omitempty"`
}

type ModifyChannelResponse struct{}

func GetChannelInformation(broadcasterIDs []string) ([]ChannelInfo, error) {
	url := HelixBaseURL + "/channels"
	if len(broadcasterIDs) > 0 {
		url = AddIDsParam(url, "broadcaster_id", broadcasterIDs)
	}

	result, err := ExecuteRequest[struct {
		Data []ChannelInfo `json:"data"`
	}]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetChannelInformation: broadcasterIDs=%v error=%v", broadcasterIDs, err)
		return nil, err
	}

	return result.Data, nil
}

func ModifyChannelInformation(req UpdateChannelRequest) error {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
	}
	url := BuildURL(HelixBaseURL+"/channels", opts)

	_, err := ExecuteJSONRequest[ModifyChannelResponse, UpdateChannelRequest]("PATCH", url, req, 204)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] ModifyChannelInformation: broadcasterID=%v error=%v", user.UserID, err)
		return err
	}

	return nil
}

func GetChannelEditors() ([]ChannelEditor, error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
	}
	url := BuildURL(HelixBaseURL+"/channels/editors", opts)

	result, err := ExecuteRequest[GetChannelEditorsResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetChannelEditors: broadcasterID=%v error=%v", user.UserID, err)
		return nil, err
	}

	return result.Data, nil
}

type FollowedChannel struct {
	BroadcasterID    string `json:"broadcaster_id"`
	BroadcasterLogin string `json:"broadcaster_login"`
	BroadcasterName  string `json:"broadcaster_name"`
	FollowedAt       string `json:"followed_at"`
}

type GetFollowedChannelsResponse struct {
	Data       []FollowedChannel `json:"data"`
	Pagination Pagination        `json:"pagination"`
}

func GetFollowedChannels(broadcasterID string, req *PaginationRequest) (*PaginationData[FollowedChannel], error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"user_id":        user.UserID,
		"broadcaster_id": broadcasterID,
	}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := BuildURL(HelixBaseURL+"/channels/followed", opts)

	result, err := ExecuteRequest[PaginationData[FollowedChannel]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetFollowedChannels: broadcasterID=%v error=%v", broadcasterID, err)
		return nil, err
	}

	quantity := 0
	if req != nil {
		quantity = req.Quantity
	}
	result.GetNext = func() *PaginationData[FollowedChannel] {
		r, _ := GetFollowedChannels(broadcasterID, &PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: quantity,
		})
		return r
	}

	return result, nil
}

func GetChannelFollowers(userId string, req *PaginationRequest) (*PaginationData[TwitchViewerData], error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"user_id":        userId,
		"broadcaster_id": user.UserID,
	}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := BuildURL(HelixBaseURL+"/channels/followers", opts)

	result, err := ExecuteRequest[PaginationData[TwitchViewerData]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetChannelFollowers: userID=%v error=%v", userId, err)
		return nil, err
	}

	quantity := 0
	if req != nil {
		quantity = req.Quantity
	}
	result.GetNext = func() *PaginationData[TwitchViewerData] {
		r, _ := GetChannelFollowers(userId, &PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: quantity,
		})
		return r
	}

	return result, nil
}
