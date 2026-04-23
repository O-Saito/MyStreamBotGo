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
	"time"
)

var urlAPIMod = "https://api.twitch.tv/helix/moderation"
var urlAPIAutoMod = "https://api.twitch.tv/helix/automod"
var urlAPIAutoModSettings = "https://api.twitch.tv/helix/automod/settings"
var urlAPIBanned = "https://api.twitch.tv/helix/moderation/banned"
var urlAPIBlockTerms = "https://api.twitch.tv/helix/moderation/blocked_terms"
var urlAPIModerators = "https://api.twitch.tv/helix/moderation/moderators"
var urlAPIVIPs = "https://api.twitch.tv/helix/moderation/vips"
var urlAPIModeratedChannels = "https://api.twitch.tv/helix/moderation/channels"
var urlAPIShieldMode = "https://api.twitch.tv/helix/moderation/shield_mode"
var urlAPIUnbanRequests = "https://api.twitch.tv/helix/moderation/unban_requests"

type AutoModStatus struct {
	Permitted bool `json:"permitted"`
}

type AutoModMessage struct {
	MsgID    string `json:"msg_id"`
	Action   string `json:"action"`
}

type AutoModSettings struct {
	OverallLevel int  `json:"overall_level"`
	AdulteContent bool `json:"adult_content"`
	Aggressive   bool `json:"aggressive"`
	BlankChat    bool `json:"blank_chat"`
	Caps         bool `json:"caps"`
	CopyPaste    bool `json:"copy_paste"`
	Custom       bool `json:"custom"`
	Emote        bool `json:"emote"`
	FollowerOnly int  `json:"follower_only"`
	Highlights   bool `json:"highlights"`
	Links        bool `json:"links"`
	Math         bool `json:"math"`
	Mention      bool `json:"mention"`
	NoSuspiciousLinks bool `json:"no_suspicious_links"`
	Profanity    bool `json:"profanity"`
	Questions    bool `json:"questions"`
	Repeated     bool `json:"repeated"`
	ShortMessages bool `json:"short_messages"`
	SlowMode     int  `json:"slow_mode"`
	Spam         bool `json:"spam"`
	Symbols      bool `json:"symbols"`
	Urls         bool `json:"urls"`
	Words        bool `json:"words"`
}

type BannedUser struct {
	UserID          string    `json:"user_id"`
	UserLogin       string    `json:"user_login"`
	UserName        string    `json:"user_name"`
	BroadcasterID   string    `json:"broadcaster_id"`
	BroadcasterLogin string   `json:"broadcaster_login"`
	BroadcasterName string    `json:"broadcaster_name"`
	ExpiresAt       time.Time `json:"expires_at"`
	CreatedAt       time.Time `json:"created_at"`
	Reason          string    `json:"reason"`
}

type BlockedTerm struct {
	ID           string `json:"id"`
	BroadcasterID string `json:"broadcaster_id"`
	Text         string `json:"text"`
	ExpiresAt    string `json:"expires_at"`
	CreatedAt    string `json:"created_at"`
}

type Moderator struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
}

type VIP struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
}

type ShieldModeStatus struct {
	IsActive    bool   `json:"is_active"`
	ActivatedAt string `json:"activated_at"`
	LastActivatedBy string `json:"last_activated_by"`
}

type UnbanRequest struct {
	ID              string `json:"id"`
	BroadcasterID   string `json:"broadcaster_id"`
	BroadcasterName string `json:"broadcaster_name"`
	UserID          string `json:"user_id"`
	UserName        string `json:"user_name"`
	UserLogin       string `json:"user_login"`
	Status          string `json:"status"`
	CreatedAt       string `json:"created_at"`
	ResolvedAt      string `json:"resolved_at"`
	ResolvedBy      string `json:"resolved_by"`
	ResolutionText  string `json:"resolution_text"`
}

type GetAutoModStatusResponse struct {
	Data []AutoModStatus `json:"data"`
}

type GetAutoModSettingsResponse struct {
	Data []AutoModSettings `json:"data"`
}

type GetBannedUsersResponse struct {
	Data       []BannedUser  `json:"data"`
	Pagination Pagination    `json:"pagination"`
}

type GetBlockedTermsResponse struct {
	Data       []BlockedTerm `json:"data"`
	Pagination Pagination    `json:"pagination"`
}

type GetModeratorsResponse struct {
	Data       []Moderator   `json:"data"`
	Pagination Pagination    `json:"pagination"`
}

type GetVIPsResponse struct {
	Data       []VIP         `json:"data"`
	Pagination Pagination    `json:"pagination"`
}

type GetShieldModeStatusResponse struct {
	Data []ShieldModeStatus `json:"data"`
}

type GetUnbanRequestsResponse struct {
	Data       []UnbanRequest `json:"data"`
	Pagination Pagination     `json:"pagination"`
}

func CheckAutoModStatus(messages []map[string]any) ([]AutoModStatus, error) {
	data, err := json.Marshal(map[string]any{"messages": messages})
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] CheckAutoModStatus json.Marshal failed: %v", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", urlAPIAutoMod+"/check", bytes.NewBuffer(data))
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] CheckAutoModStatus http.NewRequest failed: %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] CheckAutoModStatus: messages count=%v", len(messages))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] CheckAutoModStatus io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("CheckAutoModStatus: failed: %s", body)
	}

	var result GetAutoModStatusResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] CheckAutoModStatus io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] CheckAutoModStatus json.Unmarshal failed: %v", err)
		return nil, err
	}

	return result.Data, nil
}

func UpdateAutomod(userId, msgId, action string) (string, error) {
	err := ManageHeldAutoModMessages(msgId, action)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] UpdateAutomod: userId=%v, msgId=%v, action=%v", userId, msgId, action)
		return "", err
	}
	return "", nil
}

func ManageHeldAutoModMessages(msgID, action string) error {
	data := map[string]any{
		"msg_id": msgID,
		"action": action,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] ManageHeldAutoModMessages json.Marshal failed: %v", err)
		return err
	}

	req, err := http.NewRequest("POST", urlAPIAutoMod+"/message", bytes.NewBuffer(jsonData))
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] ManageHeldAutoModMessages http.NewRequest failed: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] ManageHeldAutoModMessages: msgID=%v, action=%v", msgID, action)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] ManageHeldAutoModMessages io.ReadAll failed: %v", err)
			return err
		}
		return fmt.Errorf("ManageHeldAutoModMessages: failed: %s", body)
	}

	return nil
}

func GetAutoModSettings() (*AutoModSettings, error) {
	user := globals.GetState().GetTwitchUser()

	url := fmt.Sprintf("%s?broadcaster_id=%s&moderator_id=%s", urlAPIAutoModSettings, user.UserID, user.UserID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetAutoModSettings http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetAutoModSettings: broadcasterID=%v", user.UserID)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetAutoModSettings io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("GetAutoModSettings: failed: %s", body)
	}

	var result GetAutoModSettingsResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetAutoModSettings io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetAutoModSettings json.Unmarshal failed: %v", err)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

type UpdateAutoModSettingsRequest struct {
	OverallLevel *int  `json:"overall_level,omitempty"`
	AdulteContent *bool `json:"adult_content,omitempty"`
	Aggressive   *bool `json:"aggressive,omitempty"`
	BlankChat    *bool `json:"blank_chat,omitempty"`
	Caps         *bool `json:"caps,omitempty"`
	CopyPaste    *bool `json:"copy_paste,omitempty"`
	Custom       *bool `json:"custom,omitempty"`
	Emote        *bool `json:"emote,omitempty"`
	FollowerOnly *int  `json:"follower_only,omitempty"`
	Highlights   *bool `json:"highlights,omitempty"`
	Links        *bool `json:"links,omitempty"`
	Math         *bool `json:"math,omitempty"`
	Mention      *bool `json:"mention,omitempty"`
	NoSuspiciousLinks *bool `json:"no_suspicious_links,omitempty"`
	Profanity    *bool `json:"profanity,omitempty"`
	Questions    *bool `json:"questions,omitempty"`
	Repeated     *bool `json:"repeated,omitempty"`
	ShortMessages *bool `json:"short_messages,omitempty"`
	SlowMode     *int  `json:"slow_mode,omitempty"`
	Spam         *bool `json:"spam,omitempty"`
	Symbols      *bool `json:"symbols,omitempty"`
	Urls         *bool `json:"urls,omitempty"`
	Words        *bool `json:"words,omitempty"`
}

func UpdateAutoModSettings(req UpdateAutoModSettingsRequest) error {
	user := globals.GetState().GetTwitchUser()

	data, err := json.Marshal(req)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateAutoModSettings json.Marshal failed: %v", err)
		return err
	}

	url := fmt.Sprintf("%s?broadcaster_id=%s&moderator_id=%s", urlAPIAutoModSettings, user.UserID, user.UserID)
	httpReq, err := http.NewRequest("PATCH", url, bytes.NewBuffer(data))
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateAutoModSettings http.NewRequest failed: %v", err)
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := twitch.DoRequest(httpReq)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] UpdateAutoModSettings: broadcasterID=%v", user.UserID)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] UpdateAutoModSettings io.ReadAll failed: %v", err)
			return err
		}
		return fmt.Errorf("UpdateAutoModSettings: failed: %s", body)
	}

	return nil
}

func GetBannedUsers(userIDs []string, req *twitch.PaginationRequest) (*twitch.PaginationData[BannedUser], error) {
	user := globals.GetState().GetTwitchUser()

	opts := twitch.RequestOptions{
		BroadcasterID: user.UserID,
	}

	if req != nil {
		opts.After = req.Cursor
		opts.First = req.Quantity
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/moderation/banned", opts)
	for _, id := range userIDs {
		url += fmt.Sprintf("&user_id=%s", id)
	}

	result, err := twitch.ExecuteRequest[twitch.PaginationData[BannedUser]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetBannedUsers: broadcasterID=%v", user.UserID)
		return nil, err
	}

	result.GetNext = func() *twitch.PaginationData[BannedUser] {
		GetBannedUsers(userIDs, &twitch.PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: opts.First,
		})
		return result
	}

	return result, nil
}

func UnbanUser(broadcasterID, userID string) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s?broadcaster_id=%s&moderator_id=%s&user_id=%s", urlAPIBanned, broadcasterID, user.UserID, userID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UnbanUser http.NewRequest failed: %v", err)
		return err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] UnbanUser: broadcasterID=%v, userID=%v", broadcasterID, userID)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] UnbanUser io.ReadAll failed: %v", err)
			return err
		}
		return fmt.Errorf("UnbanUser: failed: %s", body)
	}

	return nil
}

func BanUser(userId string, duration int32, reason string) (string, error) {
	user := globals.GetState().TwitchUser
	d := map[string]map[string]any{
		"data": {
			"user_id": userId,
			"reason":  reason,
		},
	}

	if duration > 0 {
		d["data"]["duration"] = duration
	}

	data, err := json.Marshal(d)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] BanUser json.Marshal failed: %v", err)
		return "", err
	}
	urlAPI := fmt.Sprintf("%s?broadcaster_id=%s&moderator_id=%s", urlAPIBanned, user.UserID, user.UserID)
	req, err := http.NewRequest("POST", urlAPI, bytes.NewBuffer(data))
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] BanUser http.NewRequest failed: %v", err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] BanUser: userId=%v, duration=%v, reason=%v", userId, duration, reason)
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] BanUser io.ReadAll failed: %v", err)
			return "", err
		}
		helpers.Logf(helpers.DEBUG, "[TWITCH] BanUser: userId=%v, duration=%v, reason=%v", userId, duration, reason)
		return "", fmt.Errorf("BanUser(%s, %d, %s): failed to ban user: %s", userId, duration, reason, body)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] BanUser io.ReadAll failed: %v", err)
		return "", err
	}
	return string(body), nil
}

func GetBlockedTerms(req *twitch.PaginationRequest) (*twitch.PaginationData[BlockedTerm], error) {
	user := globals.GetState().GetTwitchUser()

	opts := twitch.RequestOptions{
		BroadcasterID: user.UserID,
	}

	if req != nil {
		opts.After = req.Cursor
		opts.First = req.Quantity
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/moderation/blocked_terms", opts)

	result, err := twitch.ExecuteRequest[twitch.PaginationData[BlockedTerm]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetBlockedTerms: broadcasterID=%v", user.UserID)
		return nil, err
	}

	result.GetNext = func() *twitch.PaginationData[BlockedTerm] {
		GetBlockedTerms(&twitch.PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: opts.First,
		})
		return result
	}

	return result, nil
}

func AddBlockedTerm(text string, duration int) (*BlockedTerm, error) {
	user := globals.GetState().GetTwitchUser()

	data := map[string]any{
		"broadcaster_id": user.UserID,
		"moderator_id":  user.UserID,
		"text":          text,
	}
	if duration > 0 {
		data["duration_seconds"] = duration
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] AddBlockedTerm json.Marshal failed: %v", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", urlAPIBlockTerms, bytes.NewBuffer(jsonData))
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] AddBlockedTerm http.NewRequest failed: %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] AddBlockedTerm: broadcasterID=%v, text=%v", user.UserID, text)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] AddBlockedTerm io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("AddBlockedTerm: failed: %s", body)
	}

	var result struct {
		Data []BlockedTerm `json:"data"`
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] AddBlockedTerm io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] AddBlockedTerm json.Unmarshal failed: %v", err)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func RemoveBlockedTerm(termID string) error {
	user := globals.GetState().GetTwitchUser()

	url := fmt.Sprintf("%s?broadcaster_id=%s&moderator_id=%s&id=%s", urlAPIBlockTerms, user.UserID, user.UserID, termID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] RemoveBlockedTerm http.NewRequest failed: %v", err)
		return err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] RemoveBlockedTerm: broadcasterID=%v, termID=%v", user.UserID, termID)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] RemoveBlockedTerm io.ReadAll failed: %v", err)
			return err
		}
		return fmt.Errorf("RemoveBlockedTerm: failed: %s", body)
	}

	return nil
}

func GetModeratedChannels(userID string) ([]string, error) {
	url := fmt.Sprintf("%s?user_id=%s", urlAPIModeratedChannels, userID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetModeratedChannels http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetModeratedChannels: userID=%v", userID)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetModeratedChannels io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("GetModeratedChannels: failed: %s", body)
	}

	var result struct {
		Data []struct {
			BroadcasterID string `json:"broadcaster_id"`
		} `json:"data"`
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetModeratedChannels io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetModeratedChannels json.Unmarshal failed: %v", err)
		return nil, err
	}

	channels := make([]string, len(result.Data))
	for i, d := range result.Data {
		channels[i] = d.BroadcasterID
	}

	return channels, nil
}

func GetModerators(userIDs []string, req *twitch.PaginationRequest) (*twitch.PaginationData[Moderator], error) {
	user := globals.GetState().GetTwitchUser()

	opts := twitch.RequestOptions{
		BroadcasterID: user.UserID,
	}

	if req != nil {
		opts.After = req.Cursor
		opts.First = req.Quantity
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/moderation/moderators", opts)
	for _, id := range userIDs {
		url += fmt.Sprintf("&user_id=%s", id)
	}

	result, err := twitch.ExecuteRequest[twitch.PaginationData[Moderator]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetModerators: broadcasterID=%v", user.UserID)
		return nil, err
	}

	result.GetNext = func() *twitch.PaginationData[Moderator] {
		GetModerators(userIDs, &twitch.PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: opts.First,
		})
		return result
	}

	return result, nil
}

func AddChannelModerator(broadcasterID, userID string) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s?broadcaster_id=%s&user_id=%s", urlAPIModerators, broadcasterID, userID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] AddChannelModerator http.NewRequest failed: %v", err)
		return err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] AddChannelModerator: broadcasterID=%v, userID=%v", broadcasterID, userID)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] AddChannelModerator io.ReadAll failed: %v", err)
			return err
		}
		return fmt.Errorf("AddChannelModerator: failed: %s", body)
	}

	return nil
}

func RemoveChannelModerator(broadcasterID, userID string) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s?broadcaster_id=%s&user_id=%s", urlAPIModerators, broadcasterID, userID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] RemoveChannelModerator http.NewRequest failed: %v", err)
		return err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] RemoveChannelModerator: broadcasterID=%v, userID=%v", broadcasterID, userID)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] RemoveChannelModerator io.NewRequest failed: %v", err)
			return err
		}
		return fmt.Errorf("RemoveChannelModerator: failed: %s", body)
	}

	return nil
}

func GetVIPs(userIDs []string, req *twitch.PaginationRequest) (*twitch.PaginationData[VIP], error) {
	user := globals.GetState().GetTwitchUser()

	opts := twitch.RequestOptions{
		BroadcasterID: user.UserID,
	}

	if req != nil {
		opts.After = req.Cursor
		opts.First = req.Quantity
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/moderation/vips", opts)
	for _, id := range userIDs {
		url += fmt.Sprintf("&user_id=%s", id)
	}

	result, err := twitch.ExecuteRequest[twitch.PaginationData[VIP]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetVIPs: broadcasterID=%v", user.UserID)
		return nil, err
	}

	result.GetNext = func() *twitch.PaginationData[VIP] {
		GetVIPs(userIDs, &twitch.PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: opts.First,
		})
		return result
	}

	return result, nil
}

func AddChannelVIP(broadcasterID, userID string) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s?broadcaster_id=%s&user_id=%s", urlAPIVIPs, broadcasterID, userID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] AddChannelVIP http.NewRequest failed: %v", err)
		return err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] AddChannelVIP: broadcasterID=%v, userID=%v", broadcasterID, userID)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] AddChannelVIP io.ReadAll failed: %v", err)
			return err
		}
		return fmt.Errorf("AddChannelVIP: failed: %s", body)
	}

	return nil
}

func RemoveChannelVIP(broadcasterID, userID string) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s?broadcaster_id=%s&user_id=%s", urlAPIVIPs, broadcasterID, userID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] RemoveChannelVIP http.NewRequest failed: %v", err)
		return err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] RemoveChannelVIP: broadcasterID=%v, userID=%v", broadcasterID, userID)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] RemoveChannelVIP io.ReadAll failed: %v", err)
			return err
		}
		return fmt.Errorf("RemoveChannelVIP: failed: %s", body)
	}

	return nil
}

func GetShieldModeStatus() (*ShieldModeStatus, error) {
	user := globals.GetState().GetTwitchUser()

	url := fmt.Sprintf("%s?broadcaster_id=%s&moderator_id=%s", urlAPIShieldMode, user.UserID, user.UserID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetShieldModeStatus http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetShieldModeStatus: broadcasterID=%v", user.UserID)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetShieldModeStatus io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("GetShieldModeStatus: failed: %s", body)
	}

	var result GetShieldModeStatusResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetShieldModeStatus io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetShieldModeStatus json.Unmarshal failed: %v", err)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func UpdateShieldModeStatus(isActive bool) error {
	user := globals.GetState().GetTwitchUser()

	data := map[string]any{
		"is_active": isActive,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateShieldModeStatus json.Marshal failed: %v", err)
		return err
	}

	url := fmt.Sprintf("%s?broadcaster_id=%s&moderator_id=%s", urlAPIShieldMode, user.UserID, user.UserID)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateShieldModeStatus http.NewRequest failed: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] UpdateShieldModeStatus: broadcasterID=%v, isActive=%v", user.UserID, isActive)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] UpdateShieldModeStatus io.ReadAll failed: %v", err)
			return err
		}
		return fmt.Errorf("UpdateShieldModeStatus: failed: %s", body)
	}

	return nil
}

func WarnChatUser(userID, reason string) error {
	user := globals.GetState().GetTwitchUser()

	data := map[string]any{
		"broadcaster_id": user.UserID,
		"moderator_id":  user.UserID,
		"user_id":       userID,
		"reason":       reason,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] WarnChatUser json.Marshal failed: %v", err)
		return err
	}

	url := urlAPIMod + "/warns"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] WarnChatUser http.NewRequest failed: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] WarnChatUser: broadcasterID=%v, userID=%v", user.UserID, userID)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] WarnChatUser io.ReadAll failed: %v", err)
			return err
		}
		return fmt.Errorf("WarnChatUser: failed: %s", body)
	}

	return nil
}

func AddSuspiciousStatusToChatUser(broadcasterID, moderatorID, userID string) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}
	if moderatorID == "" {
		moderatorID = user.UserID
	}

	data := map[string]any{
		"broadcaster_id": broadcasterID,
		"moderator_id":   moderatorID,
		"user_id":        userID,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] AddSuspiciousStatusToChatUser json.Marshal failed: %v", err)
		return err
	}

	url := urlAPIMod + "/suspicious"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] AddSuspiciousStatusToChatUser http.NewRequest failed: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] AddSuspiciousStatusToChatUser: broadcasterID=%v, userID=%v", broadcasterID, userID)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] AddSuspiciousStatusToChatUser io.ReadAll failed: %v", err)
			return err
		}
		return fmt.Errorf("AddSuspiciousStatusToChatUser: failed: %s", body)
	}

	return nil
}

func RemoveSuspiciousStatusFromChatUser(broadcasterID, moderatorID, userID string) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}
	if moderatorID == "" {
		moderatorID = user.UserID
	}

	url := fmt.Sprintf("%s/suspicious?broadcaster_id=%s&moderator_id=%s&user_id=%s", urlAPIMod, broadcasterID, moderatorID, userID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] RemoveSuspiciousStatusFromChatUser http.NewRequest failed: %v", err)
		return err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] RemoveSuspiciousStatusFromChatUser: broadcasterID=%v, userID=%v", broadcasterID, userID)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] RemoveSuspiciousStatusFromChatUser io.ReadAll failed: %v", err)
			return err
		}
		return fmt.Errorf("RemoveSuspiciousStatusFromChatUser: failed: %s", body)
	}

	return nil
}

func GetUnbanRequests(userID, status string, req *twitch.PaginationRequest) (*twitch.PaginationData[UnbanRequest], error) {
	user := globals.GetState().GetTwitchUser()

	opts := twitch.RequestOptions{
		BroadcasterID: user.UserID,
		UserID:      userID,
	}

	if req != nil {
		opts.After = req.Cursor
		opts.First = req.Quantity
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/moderation/unban_requests", opts)
	if status != "" {
		url += fmt.Sprintf("&status=%s", status)
	}

	result, err := twitch.ExecuteRequest[twitch.PaginationData[UnbanRequest]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetUnbanRequests: broadcasterID=%v", user.UserID)
		return nil, err
	}

	result.GetNext = func() *twitch.PaginationData[UnbanRequest] {
		GetUnbanRequests(userID, status, &twitch.PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: opts.First,
		})
		return result
	}

	return result, nil
}

func ResolveUnbanRequest(requestID, action, resolutionText string) error {
	user := globals.GetState().GetTwitchUser()

	data := map[string]any{
		"status":          action,
		"resolution_text": resolutionText,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] ResolveUnbanRequest json.Marshal failed: %v", err)
		return err
	}

	url := fmt.Sprintf("%s?broadcaster_id=%s&id=%s", urlAPIUnbanRequests, user.UserID, requestID)
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] ResolveUnbanRequest http.NewRequest failed: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] ResolveUnbanRequest: broadcasterID=%v, requestID=%v, action=%v", user.UserID, requestID, action)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] ResolveUnbanRequest io.ReadAll failed: %v", err)
			return err
		}
		return fmt.Errorf("ResolveUnbanRequest: failed: %s", body)
	}

	return nil
}