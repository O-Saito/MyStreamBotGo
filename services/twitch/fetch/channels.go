package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
)

var (
	urlAPIChannels         = twitch.HelixBaseURL + "/channels"
	urlAPIChannelEditors   = twitch.HelixBaseURL + "/channels/editors"
	urlAPIFollowedChannels = twitch.HelixBaseURL + "/channels/followed"
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
	Tags                        []string `json:"tags"`
	ContentClassificationLabels []string `json:"content_classification_labels"`
}

type GetChannelEditorsResponse struct {
	Data []ChannelEditor `json:"data"`
}

type UpdateChannelLabel struct {
	// ID is one of (DebatedSocialIssuesAndPolitics, DrugsIntoxication, SexualThemes, ViolentGraphic, Gambling, ProfanityVulgarity)
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

// TODO: Check if has at least one broadcasterID
func GetChannelInformation(broadcasterIDs []string) ([]ChannelInfo, error) {
	url := twitch.HelixBaseURL + "/channels"
	if len(broadcasterIDs) > 0 {
		url = twitch.AddIDsParam(url, "broadcaster_id", broadcasterIDs)
	}

	result, err := twitch.ExecuteRequest[struct {
		Data []ChannelInfo `json:"data"`
	}]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetChannelInformation: broadcasterIDs=%v", broadcasterIDs)
		return nil, err
	}

	return result.Data, nil
}

func ModifyChannelInformation(req UpdateChannelRequest) error {
	user := globals.GetState().GetTwitchUser()

	url := twitch.AddIDParam(twitch.HelixBaseURL+"/channels", "broadcaster_id", user.UserID)

	_, err := twitch.ExecuteJSONRequest[ModifyChannelResponse]("PATCH", url, req, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] ModifyChannelInformation: broadcasterID=%v", user.UserID)
		return err
	}

	return nil
}

func GetChannelEditors() ([]ChannelEditor, error) {
	user := globals.GetState().GetTwitchUser()

	url := twitch.AddIDParam(twitch.HelixBaseURL+"/channels/editors", "broadcaster_id", user.UserID)

	result, err := twitch.ExecuteRequest[GetChannelEditorsResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetChannelEditors: broadcasterID=%v", user.UserID)
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
	Pagination twitch.Pagination `json:"pagination"`
}

func GetFollowedChannels(broadcaster_id string, first int, after string) ([]FollowedChannel, error) {
	user := globals.GetState().GetTwitchUser()

	opts := twitch.RequestOptions{
		UserID:        user.UserID,
		BroadcasterID: broadcaster_id,
		First:         first,
		After:         after,
	}
	url := twitch.BuildURL(twitch.HelixBaseURL+"/channels/followed", opts)

	result, err := twitch.ExecuteRequest[GetFollowedChannelsResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetFollowedChannels: userID=%v", user.UserID)
		return nil, err
	}

	return result.Data, nil
}
