package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var urlAPIChat = "https://api.twitch.tv/helix/chat"
var urlAPIChatAnnouncements = "https://api.twitch.tv/helix/chat/announcements"
var urlAPIChatShoutouts = "https://api.twitch.tv/helix/chat/shoutouts"
var urlAPIChatSettings = "https://api.twitch.tv/helix/chat/settings"
var urlAPIChatColor = "https://api.twitch.tv/helix/chat/color"
var urlAPIChatters = "https://api.twitch.tv/helix/chat/chatters"
var urlAPIEmotes = "https://api.twitch.tv/helix/chat/emotes"
var urlAPIEmotesSets = "https://api.twitch.tv/helix/chat/emotes/set"
var urlAPIChatBadges = "https://api.twitch.tv/helix/chat/badges"
var urlAPISharedChat = "https://api.twitch.tv/helix/chat/shared_chat"

type ChatSettings struct {
	EmoteOnly             bool   `json:"emote_only"`
	FollowersOnly         int    `json:"followers_only"`
	NonModeratorChatDelay int    `json:"non_moderator_chat_delay"`
	NonModeratorChatDelayDuration int `json:"non_moderator_chat_delay_duration"`
	R9K                   bool   `json:"r9k"`
	Slow                  int    `json:"slow"`
	SubsOnly              bool   `json:"subs_only"`
}

type ChannelEmote struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Images    EmoteImages `json:"images"`
	Tier      string `json:"tier"`
	Type      string `json:"type"`
	Format    []string `json:"format"`
	Scale     []string `json:"scale"`
	ThemeMode []string `json:"theme_mode"`
}

type EmoteImages struct {
	URL1x string `json:"url_1x"`
	URL2x string `json:"url_2x"`
	URL4x string `json:"url_4x"`
}

type GlobalEmote struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Images    EmoteImages `json:"images"`
	Type      string      `json:"type"`
	Format    []string    `json:"format"`
	Scale     []string    `json:"scale"`
	ThemeMode []string    `json:"theme_mode"`
}

type EmoteSet struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Images EmoteImages `json:"images"`
}

type ChatBadgeVersion struct {
	ID          int    `json:"id"`
	ImageURL1x  string `json:"image_url_1x"`
	ImageURL2x  string `json:"image_url_2x"`
	ImageURL4x  string `json:"image_url_4x"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type ChatBadgeSet struct {
	SetID    string             `json:"set_id"`
	Versions []ChatBadgeVersion `json:"versions"`
}

type Chatter struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
}

type GetChattersResponse struct {
	Data       []Chatter `json:"data"`
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

func GetChatters(broadcasterID, moderatorID string, req *twitch.PaginationRequest) (*twitch.PaginationData[Chatter], error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}
	if moderatorID == "" {
		moderatorID = user.UserID
	}

	opts := twitch.RequestOptions{
		BroadcasterID: broadcasterID,
		ModeratorID:  moderatorID,
	}

	if req != nil {
		opts.After = req.Cursor
		opts.First = req.Quantity
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/chat/chatters", opts)

	result, err := twitch.ExecuteRequest[twitch.PaginationData[Chatter]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetChatters: broadcasterID=%v", broadcasterID)
		return nil, err
	}

	result.GetNext = func() *twitch.PaginationData[Chatter] {
		GetChatters(broadcasterID, moderatorID, &twitch.PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: opts.First,
		})
		return result
	}

	return result, nil
}

func GetChannelEmotes(broadcasterID string) ([]ChannelEmote, error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s?broadcaster_id=%s", urlAPIEmotes, broadcasterID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetChannelEmotes http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetChannelEmotes: broadcasterID=%v", broadcasterID)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetChannelEmotes io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("GetChannelEmotes: failed: %s", body)
	}

	var result GetChannelEmotesResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetChannelEmotes io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetChannelEmotes json.Unmarshal failed: %v", err)
		return nil, err
	}

	return result.Data, nil
}

func GetGlobalEmotes() ([]GlobalEmote, error) {
	req, err := http.NewRequest("GET", urlAPIEmotes, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetGlobalEmotes http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetGlobalEmotes: no params")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetGlobalEmotes io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("GetGlobalEmotes: failed: %s", body)
	}

	var result GetGlobalEmotesResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetGlobalEmotes io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetGlobalEmotes json.Unmarshal failed: %v", err)
		return nil, err
	}

	return result.Data, nil
}

func GetEmoteSets(emoteSetIDs []string) ([]EmoteSet, error) {
	url := urlAPIEmotesSets + "?emote_set_id=" + fmt.Sprintf("&emote_set_id=", emoteSetIDs[0])
	for i := 1; i < len(emoteSetIDs); i++ {
		url += fmt.Sprintf("&emote_set_id=%s", emoteSetIDs[i])
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetEmoteSets http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetEmoteSets: emoteSetIDs=%v", emoteSetIDs)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetEmoteSets io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("GetEmoteSets: failed: %s", body)
	}

	var result GetEmoteSetsResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetEmoteSets io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetEmoteSets json.Unmarshal failed: %v", err)
		return nil, err
	}

	return result.Data, nil
}

func GetChannelChatBadges(broadcasterID string) ([]ChatBadgeSet, error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s?broadcaster_id=%s", urlAPIChatBadges, broadcasterID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetChannelChatBadges http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetChannelChatBadges: broadcasterID=%v", broadcasterID)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetChannelChatBadges io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("GetChannelChatBadges: failed: %s", body)
	}

	var result GetChatBadgesResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetChannelChatBadges io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetChannelChatBadges json.Unmarshal failed: %v", err)
		return nil, err
	}

	return result.Data, nil
}

func GetGlobalChatBadges() ([]ChatBadgeSet, error) {
	url := urlAPIChatBadges + "/global"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetGlobalChatBadges http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetGlobalChatBadges: no params")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetGlobalChatBadges io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("GetGlobalChatBadges: failed: %s", body)
	}

	var result GetChatBadgesResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetGlobalChatBadges io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetGlobalChatBadges json.Unmarshal failed: %v", err)
		return nil, err
	}

	return result.Data, nil
}

func GetChatSettings(broadcasterID, moderatorID string) (*ChatSettings, error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}
	if moderatorID == "" {
		moderatorID = user.UserID
	}

	url := fmt.Sprintf("%s?broadcaster_id=%s&moderator_id=%s", urlAPIChatSettings, broadcasterID, moderatorID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetChatSettings http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetChatSettings: broadcasterID=%v", broadcasterID)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetChatSettings io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("GetChatSettings: failed: %s", body)
	}

	var result GetChatSettingsResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetChatSettings io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetChatSettings json.Unmarshal failed: %v", err)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

type UpdateChatSettingsRequest struct {
	EmoteOnly             *bool `json:"emote_only,omitempty"`
	FollowersOnly         *int  `json:"followers_only,omitempty"`
	NonModeratorChatDelay *bool `json:"non_moderator_chat_delay,omitempty"`
	NonModeratorChatDelayDuration *int `json:"non_moderator_chat_delay_duration,omitempty"`
	R9K                   *bool `json:"r9k,omitempty"`
	Slow                  *int  `json:"slow,omitempty"`
	SubsOnly              *bool `json:"subs_only,omitempty"`
}

func UpdateChatSettings(broadcasterID, moderatorID string, req UpdateChatSettingsRequest) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}
	if moderatorID == "" {
		moderatorID = user.UserID
	}

	data, err := json.Marshal(req)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateChatSettings json.Marshal failed: %v", err)
		return err
	}

	url := fmt.Sprintf("%s?broadcaster_id=%s&moderator_id=%s", urlAPIChatSettings, broadcasterID, moderatorID)
	httpReq, err := http.NewRequest("PATCH", url, bytes.NewBuffer(data))
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateChatSettings http.NewRequest failed: %v", err)
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := twitch.DoRequest(httpReq)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] UpdateChatSettings: broadcasterID=%v", broadcasterID)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] UpdateChatSettings io.ReadAll failed: %v", err)
			return err
		}
		return fmt.Errorf("UpdateChatSettings: failed: %s", body)
	}

	return nil
}

func SendChatAnnouncement(broadcasterID, moderatorID, message, color string) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}
	if moderatorID == "" {
		moderatorID = user.UserID
	}

	data := map[string]any{
		"message":      message,
		"broadcaster_id": broadcasterID,
		"moderator_id": moderatorID,
	}
	if color != "" {
		data["color"] = color
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] SendChatAnnouncement json.Marshal failed: %v", err)
		return err
	}

	req, err := http.NewRequest("POST", urlAPIChatAnnouncements, bytes.NewBuffer(jsonData))
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] SendChatAnnouncement http.NewRequest failed: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] SendChatAnnouncement: broadcasterID=%v, message=%v", broadcasterID, message)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] SendChatAnnouncement io.ReadAll failed: %v", err)
			return err
		}
		return fmt.Errorf("SendChatAnnouncement: failed: %s", body)
	}

	return nil
}

func SendShoutout(fromBroadcasterID, toBroadcasterID, moderatorID string) error {
	user := globals.GetState().GetTwitchUser()
	if fromBroadcasterID == "" {
		fromBroadcasterID = user.UserID
	}
	if moderatorID == "" {
		moderatorID = user.UserID
	}

	url := fmt.Sprintf("%s?from_broadcaster_id=%s&to_broadcaster_id=%s&moderator_id=%s", 
		urlAPIChatShoutouts, fromBroadcasterID, toBroadcasterID, moderatorID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] SendShoutout http.NewRequest failed: %v", err)
		return err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] SendShoutout: fromBroadcasterID=%v, toBroadcasterID=%v", fromBroadcasterID, toBroadcasterID)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] SendShoutout io.ReadAll failed: %v", err)
			return err
		}
		return fmt.Errorf("SendShoutout: failed: %s", body)
	}

	return nil
}

type SendChatMessageRequest struct {
	Message      string `json:"message"`
	ReplyParentMessageID string `json:"reply_parent_message_id,omitempty"`
}

func SendChatMessage(broadcasterID, moderatorID string, req SendChatMessageRequest) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}
	if moderatorID == "" {
		moderatorID = user.UserID
	}

	data, err := json.Marshal(req)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] SendChatMessage json.Marshal failed: %v", err)
		return err
	}

	url := fmt.Sprintf("%s?broadcaster_id=%s&moderator_id=%s", urlAPIChat, broadcasterID, moderatorID)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] SendChatMessage http.NewRequest failed: %v", err)
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := twitch.DoRequest(httpReq)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] SendChatMessage: broadcasterID=%v, message=%v", broadcasterID, req.Message)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] SendChatMessage io.ReadAll failed: %v", err)
			return err
		}
		return fmt.Errorf("SendChatMessage: failed: %s", body)
	}

	return nil
}

func UpdateUserChatColor(userID, color string) error {
	user := globals.GetState().GetTwitchUser()
	if userID == "" {
		userID = user.UserID
	}

	url := fmt.Sprintf("%s?user_id=%s&color=%s", urlAPIChatColor, userID, color)
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateUserChatColor http.NewRequest failed: %v", err)
		return err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] UpdateUserChatColor: userID=%v, color=%v", userID, color)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] UpdateUserChatColor io.ReadAll failed: %v", err)
			return err
		}
		return fmt.Errorf("UpdateUserChatColor: failed: %s", body)
	}

	return nil
}

func DeleteMessage(msgID string) error {
	user := globals.GetState().TwitchUser
	urlAPI := fmt.Sprintf("https://api.twitch.tv/helix/moderation/chat?broadcaster_id=%s&moderator_id=%s&message_id=%s", user.UserID, user.UserID, msgID)
	req, err := http.NewRequest("DELETE", urlAPI, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] DeleteMessage http.NewRequest failed: %v", err)
		return err
	}
	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] DeleteMessage: msgID=%v", msgID)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] DeleteMessage io.ReadAll failed: %v", err)
			return err
		}
		helpers.Logf(helpers.DEBUG, "[TWITCH] DeleteMessage: msgID=%v", msgID)
		return fmt.Errorf("DeleteMessage(%s): failed to delete message: %s", msgID, body)
	}
	return nil
}

type SharedChatSession struct {
	SharedChatEnabled bool   `json:"shared_chat_enabled"`
	Color           string `json:"color"`
	EmoteSetID      string `json:"emote_set_id"`
	DisplayName     string `json:"display_name"`
	Login           string `json:"login"`
	UserID          string `json:"user_id"`
}

type GetSharedChatSessionResponse struct {
	Data []SharedChatSession `json:"data"`
}

func GetSharedChatSession(broadcasterID, userID string) (*SharedChatSession, error) {
	user := globals.GetState().GetTwitchUser()
	if userID == "" {
		userID = user.UserID
	}
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s?broadcaster_id=%s&user_id=%s", urlAPISharedChat, broadcasterID, userID)
	resp, err := twitch.ExecuteRequest[GetSharedChatSessionResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetSharedChatSession: broadcasterID=%v, userID=%v", broadcasterID, userID)
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

type UserEmote struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Images   EmoteImages `json:"images"`
	Tier     string `json:"tier"`
	Format   []string `json:"format"`
	Scale    []string `json:"scale"`
	ThemeMode []string `json:"theme_mode"`
}

type GetUserEmotesResponse struct {
	Data []UserEmote `json:"data"`
}

func GetUserEmotes(userID string) ([]UserEmote, error) {
	if userID == "" {
		user := globals.GetState().GetTwitchUser()
		userID = user.UserID
	}

	url := fmt.Sprintf("%s?user_id=%s", urlAPIEmotes, userID)
	resp, err := twitch.ExecuteRequest[GetUserEmotesResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetUserEmotes: userID=%v", userID)
		return nil, err
	}

	return resp.Data, nil
}

func GetUserChatColor(userID string) (map[string]string, error) {
	if userID == "" {
		user := globals.GetState().GetTwitchUser()
		userID = user.UserID
	}

	url := fmt.Sprintf("%s?user_id=%s", urlAPIChatColor, userID)
	resp, err := twitch.ExecuteRequest[struct {
		Data []struct {
			UserID    string `json:"user_id"`
			UserName  string `json:"user_name"`
			UserLogin string `json:"user_login"`
			Color    string `json:"color"`
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