package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	
)

var urlAPIChat = "https://api.tv/helix/chat"
var urlAPIChatAnnouncements = "https://api.tv/helix/chat/announcements"
var urlAPIChatShoutouts = "https://api.tv/helix/chat/shoutouts"
var urlAPIChatSettings = "https://api.tv/helix/chat/settings"
var urlAPIChatColor = "https://api.tv/helix/chat/color"
var urlAPIChatters = "https://api.tv/helix/chat/chatters"
var urlAPIEmotes = "https://api.tv/helix/chat/emotes"
var urlAPIEmotesSets = "https://api.tv/helix/chat/emotes/set"
var urlAPIChatBadges = "https://api.tv/helix/chat/badges"
var urlAPISharedChat = "https://api.tv/helix/chat/shared_chat"

type ChatSettings struct {
	EmoteOnly                      bool    `json:"emote_only"`
	FollowersOnly                  int     `json:"followers_only"`
	NonModeratorChatDelay          bool    `json:"non_moderator_chat_delay"`
	NonModeratorChatDelayDuration  int     `json:"non_moderator_chat_delay_duration"`
	R9K                           bool    `json:"r9k"`
	Slow                          int     `json:"slow"`
	SubsOnly                      bool    `json:"subs_only"`
}

type ChannelEmote struct {
	ID       string         `json:"id"`
	Name    string         `json:"name"`
	Images  EmoteImages   `json:"images"`
	Tier   string         `json:"tier"`
	Type   string         `json:"type"`
	Format []string       `json:"format"`
	Scale  []string       `json:"scale"`
}

type EmoteImages struct {
	URL1x string `json:"url_1x"`
	URL2x string `json:"url_2x"`
	URL4x string `json:"url_4x"`
}

type GlobalEmote struct {
	ID       string         `json:"id"`
	Name    string         `json:"name"`
	Images  EmoteImages   `json:"images"`
	Type    string         `json:"type"`
	Format []string       `json:"format"`
	Scale  []string       `json:"scale"`
}

type EmoteSet struct {
	ID     string       `json:"id"`
	Name   string       `json:"name"`
	Images EmoteImages `json:"images"`
}

type ChatBadgeVersion struct {
	ID          int    `json:"id"`
	ImageURL1x  string `json:"image_url_1x"`
	ImageURL2x  string `json:"image_url_2x"`
	ImageURL4x  string `json:"image_url_4x"`
	Title      string `json:"title"`
}

type ChatBadgeSet struct {
	SetID    string            `json:"set_id"`
	Versions []ChatBadgeVersion `json:"versions"`
}

type Chatter struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
}

type GetChattersResponse struct {
	Data       []Chatter        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

type GetChannelEmotesResponse struct {
	Data []ChannelEmote `json:"data"`
}

type GetGlobalEmotesResponse struct {
	Data []GlobalEmote `json:"data"`
}

type GetEmoteSetsResponse struct {
	Data []EmoteSet `json:"data"`
}

type GetChatBadgesResponse struct {
	Data []ChatBadgeSet `json:"data"`
}

type GetChatSettingsResponse struct {
	Data []ChatSettings `json:"data"`
}

func GetChatters(req *PaginationRequest) (*PaginationData[Chatter], error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
		"moderator_id":   user.UserID,
	}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := BuildURL(HelixBaseURL+"/chat/chatters", opts)

	result, err := ExecuteRequest[PaginationData[Chatter]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetChatters: broadcasterID=%v error=%v", user.UserID, err)
		return nil, err
	}

	quantity := 0
	if req != nil {
		quantity = req.Quantity
	}
	result.GetNext = func() *PaginationData[Chatter] {
		r, _ := GetChatters(&PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: quantity,
		})
		return r
	}

	return result, nil
}

func GetChannelEmotes() ([]ChannelEmote, error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
	}

	url := BuildURL(HelixBaseURL+"/chat/emotes", opts)

	result, err := ExecuteRequest[GetChannelEmotesResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetChannelEmotes: broadcasterID=%v error=%v", user.UserID, err)
		return nil, err
	}

	return result.Data, nil
}

func GetGlobalEmotes() ([]GlobalEmote, error) {
	opts := map[string]any{}

	url := BuildURL(urlAPIEmotes, opts)

	result, err := ExecuteRequest[GetGlobalEmotesResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetGlobalEmotes: error=%v", err)
		return nil, err
	}

	return result.Data, nil
}

func GetEmoteSets(emoteSetIDs []string) ([]EmoteSet, error) {
	opts := map[string]any{}
	for _, id := range emoteSetIDs {
		opts["emote_set_id"] = id
	}

	url := BuildURL(urlAPIEmotesSets, opts)

	result, err := ExecuteRequest[GetEmoteSetsResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetEmoteSets: emoteSetIDs=%v error=%v", emoteSetIDs, err)
		return nil, err
	}

	return result.Data, nil
}

func GetChannelChatBadges() ([]ChatBadgeSet, error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
	}

	url := BuildURL(urlAPIChatBadges, opts)

	result, err := ExecuteRequest[GetChatBadgesResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetChannelChatBadges: broadcasterID=%v error=%v", user.UserID, err)
		return nil, err
	}

	return result.Data, nil
}

func GetGlobalChatBadges() ([]ChatBadgeSet, error) {
	opts := map[string]any{}

	url := BuildURL(urlAPIChatBadges+"/global", opts)

	result, err := ExecuteRequest[GetChatBadgesResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetGlobalChatBadges: error=%v", err)
		return nil, err
	}

	return result.Data, nil
}

func GetChatSettings() (*ChatSettings, error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
		"moderator_id":   user.UserID,
	}

	url := BuildURL(urlAPIChatSettings, opts)

	result, err := ExecuteRequest[GetChatSettingsResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetChatSettings: broadcasterID=%v error=%v", user.UserID, err)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

type UpdateChatSettingsRequest struct {
	EmoteOnly                    *bool `json:"emote_only,omitempty"`
	FollowersOnly                *int  `json:"followers_only,omitempty"`
	NonModeratorChatDelay        *bool `json:"non_moderator_chat_delay,omitempty"`
	NonModeratorChatDelayDuration *int `json:"non_moderator_chat_delay_duration,omitempty"`
	R9K                         *bool `json:"r9k,omitempty"`
	Slow                        *int  `json:"slow,omitempty"`
	SubsOnly                    *bool `json:"subs_only,omitempty"`
}

func UpdateChatSettings(req UpdateChatSettingsRequest) error {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
		"moderator_id":   user.UserID,
	}

	url := BuildURL(urlAPIChatSettings, opts)

	_, err := ExecuteJSONRequest[struct{}, UpdateChatSettingsRequest]("PATCH", url, req, 200)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateChatSettings: broadcasterID=%v error=%v", user.UserID, err)
		return err
	}

	return nil
}

func SendChatAnnouncement(message, color string) error {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
		"moderator_id":   user.UserID,
	}
	url := BuildURL(HelixBaseURL+"/chat/announcements", opts)

	body := map[string]any{
		"message": message,
	}
	if color != "" {
		body["color"] = color
	}

	_, err := ExecuteJSONRequest[struct{}, map[string]any]("POST", url, body, 204)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] SendChatAnnouncement: broadcasterID=%v, message=%v error=%v", user.UserID, message, err)
		return err
	}

	return nil
}

func SendShoutout(toBroadcasterID string) error {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"from_broadcaster_id": user.UserID,
		"to_broadcaster_id":  toBroadcasterID,
		"moderator_id":       user.UserID,
	}

	url := BuildURL(urlAPIChatShoutouts, opts)

	_, err := ExecuteRequest[struct{}]("POST", url, 204)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] SendShoutout: fromBroadcasterID=%v, toBroadcasterID=%v error=%v", user.UserID, toBroadcasterID, err)
		return err
	}

	return nil
}

type SendChatMessageRequest struct {
	Message             string `json:"message"`
	ReplyParentMessageID string `json:"reply_parent_message_id,omitempty"`
}

func SendChatMessage(msg string) error {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
		"moderator_id":  user.UserID,
	}

	url := BuildURL(HelixBaseURL+"/chat", opts)

	data := map[string]any{"message": msg}
	_, err := ExecuteJSONRequest[struct{}, map[string]any]("POST", url, data, 200)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] SendChatMessage: broadcasterID=%v error=%v", user.UserID, err)
		return err
	}

	return nil
}

func UpdateUserChatColor(userID, color string) error {
	user := globals.GetState().GetTwitchUser()
	if userID == "" {
		userID = user.UserID
	}

	opts := map[string]any{
		"user_id": userID,
		"color":  color,
	}

	url := BuildURL(urlAPIChatColor, opts)

	_, err := ExecuteRequest[struct{}]("PUT", url, 204)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateUserChatColor: userID=%v, color=%v error=%v", userID, color, err)
		return err
	}

	return nil
}

func DeleteMessage(msgID string) error {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
		"moderator_id":   user.UserID,
		"message_id":    msgID,
	}

	url := BuildURL(HelixBaseURL+"/moderation/chat", opts)

	_, err := ExecuteRequest[struct{}]("DELETE", url, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] DeleteMessage: msgID=%v", msgID)
		return err
	}

	return nil
}

type SharedChatSession struct {
	SharedChatEnabled bool   `json:"shared_chat_enabled"`
	Color          string `json:"color"`
	EmoteSetID     string `json:"emote_set_id"`
	DisplayName   string `json:"display_name"`
	Login        string `json:"login"`
	UserID       string `json:"user_id"`
}

type GetSharedChatSessionResponse struct {
	Data []SharedChatSession `json:"data"`
}

func GetSharedChatSession(userID string) (*SharedChatSession, error) {
	user := globals.GetState().GetTwitchUser()
	if userID == "" {
		userID = user.UserID
	}

	opts := map[string]any{
		"broadcaster_id": user.UserID,
		"user_id":      userID,
	}

	url := BuildURL(urlAPISharedChat, opts)

	resp, err := ExecuteRequest[GetSharedChatSessionResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetSharedChatSession: broadcasterID=%v, userID=%v", user.UserID, userID)
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

type UserEmote struct {
	ID        string       `json:"id"`
	Name     string       `json:"name"`
	Images  EmoteImages `json:"images"`
	Tier    string       `json:"tier"`
	Format  []string     `json:"format"`
	Scale   []string     `json:"scale"`
}

type GetUserEmotesResponse struct {
	Data []UserEmote `json:"data"`
}

func GetUserEmotes(userID string) ([]UserEmote, error) {
	user := globals.GetState().GetTwitchUser()
	if userID == "" {
		userID = user.UserID
	}

	opts := map[string]any{
		"user_id": userID,
	}

	url := BuildURL(urlAPIEmotes, opts)

	resp, err := ExecuteRequest[GetUserEmotesResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetUserEmotes: userID=%v", userID)
		return nil, err
	}

	return resp.Data, nil
}

func GetUserChatColor(userID string) (map[string]string, error) {
	user := globals.GetState().GetTwitchUser()
	if userID == "" {
		userID = user.UserID
	}

	opts := map[string]any{
		"user_id": userID,
	}

	url := BuildURL(urlAPIChatColor, opts)

	resp, err := ExecuteRequest[struct {
		Data []struct {
			UserID    string `json:"user_id"`
			UserName  string `json:"user_name"`
			UserLogin string `json:"user_login"`
			Color   string `json:"color"`
		} `json:"data"`
	}]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetUserChatColor: userID=%v", userID)
		return nil, err
	}

	result := make(map[string]string)
	for _, c := range resp.Data {
		result[c.UserID] = c.Color
	}
	return result, nil
}