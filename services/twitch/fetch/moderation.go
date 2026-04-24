package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	
	"time"
)

var urlAPIMod = "https://api.tv/helix/moderation"
var urlAPIAutoMod = "https://api.tv/helix/automod"
var urlAPIAutoModSettings = "https://api.tv/helix/automod/settings"
var urlAPIBanned = "https://api.tv/helix/moderation/banned"
var urlAPIBlockTerms = "https://api.tv/helix/moderation/blocked_terms"
var urlAPIModerators = "https://api.tv/helix/moderation/moderators"
var urlAPIVIPs = "https://api.tv/helix/moderation/vips"
var urlAPIModeratedChannels = "https://api.tv/helix/moderation/channels"
var urlAPIShieldMode = "https://api.tv/helix/moderation/shield_mode"
var urlAPIUnbanRequests = "https://api.tv/helix/moderation/unban_requests"

type AutoModStatus struct {
	Permitted bool `json:"permitted"`
}

type AutoModMessage struct {
	MsgID  string `json:"msg_id"`
	Action string `json:"action"`
}

type AutoModSettings struct {
	OverallLevel      int  `json:"overall_level"`
	AdulteContent     bool `json:"adult_content"`
	Aggressive        bool `json:"aggressive"`
	BlankChat         bool `json:"blank_chat"`
	Caps              bool `json:"caps"`
	CopyPaste         bool `json:"copy_paste"`
	Custom            bool `json:"custom"`
	Emote             bool `json:"emote"`
	FollowerOnly      int  `json:"follower_only"`
	Highlights        bool `json:"highlights"`
	Links             bool `json:"links"`
	Math              bool `json:"math"`
	Mention           bool `json:"mention"`
	NoSuspiciousLinks bool `json:"no_suspicious_links"`
	Profanity         bool `json:"profanity"`
	Questions         bool `json:"questions"`
	Repeated          bool `json:"repeated"`
	ShortMessages     bool `json:"short_messages"`
	SlowMode          int  `json:"slow_mode"`
	Spam              bool `json:"spam"`
	Symbols           bool `json:"symbols"`
	Urls              bool `json:"urls"`
	Words             bool `json:"words"`
}

type BannedUser struct {
	UserID           string    `json:"user_id"`
	UserLogin        string    `json:"user_login"`
	UserName         string    `json:"user_name"`
	BroadcasterID    string    `json:"broadcaster_id"`
	BroadcasterLogin string    `json:"broadcaster_login"`
	BroadcasterName  string    `json:"broadcaster_name"`
	ExpiresAt        time.Time `json:"expires_at"`
	CreatedAt        time.Time `json:"created_at"`
	Reason           string    `json:"reason"`
}

type BlockedTerm struct {
	ID            string `json:"id"`
	BroadcasterID string `json:"broadcaster_id"`
	Text          string `json:"text"`
	ExpiresAt     string `json:"expires_at"`
	CreatedAt     string `json:"created_at"`
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
	IsActive        bool   `json:"is_active"`
	ActivatedAt     string `json:"activated_at"`
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

type GetShieldModeStatusResponse struct {
	Data []ShieldModeStatus `json:"data"`
}

func CheckAutoModStatus(messages []map[string]any) ([]AutoModStatus, error) {
	url := BuildURL(HelixBaseURL+"/automod/check", nil)

	body := map[string]any{
		"messages": messages,
	}

	result, err := ExecuteJSONRequest[GetAutoModStatusResponse, map[string]any]("POST", url, body, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] CheckAutoModStatus: messages count=%v", len(messages))
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
	url := BuildURL(HelixBaseURL+"/automod/message", nil)

	body := map[string]any{
		"msg_id": msgID,
		"action": action,
	}

	_, err := ExecuteJSONRequest[map[string]any, map[string]any]("POST", url, body, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] ManageHeldAutoModMessages: msgID=%v, action=%v", msgID, action)
		return err
	}

	return nil
}

func GetAutoModSettings() (*AutoModSettings, error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
		"moderator_id":   user.UserID,
	}
	url := BuildURL(HelixBaseURL+"/automod/settings", opts)

	result, err := ExecuteRequest[GetAutoModSettingsResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetAutoModSettings: broadcasterID=%v", user.UserID)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

type UpdateAutoModSettingsRequest struct {
	OverallLevel      *int  `json:"overall_level,omitempty"`
	AdulteContent     *bool `json:"adult_content,omitempty"`
	Aggressive        *bool `json:"aggressive,omitempty"`
	BlankChat         *bool `json:"blank_chat,omitempty"`
	Caps              *bool `json:"caps,omitempty"`
	CopyPaste         *bool `json:"copy_paste,omitempty"`
	Custom            *bool `json:"custom,omitempty"`
	Emote             *bool `json:"emote,omitempty"`
	FollowerOnly      *int  `json:"follower_only,omitempty"`
	Highlights        *bool `json:"highlights,omitempty"`
	Links             *bool `json:"links,omitempty"`
	Math              *bool `json:"math,omitempty"`
	Mention           *bool `json:"mention,omitempty"`
	NoSuspiciousLinks *bool `json:"no_suspicious_links,omitempty"`
	Profanity         *bool `json:"profanity,omitempty"`
	Questions         *bool `json:"questions,omitempty"`
	Repeated          *bool `json:"repeated,omitempty"`
	ShortMessages     *bool `json:"short_messages,omitempty"`
	SlowMode          *int  `json:"slow_mode,omitempty"`
	Spam              *bool `json:"spam,omitempty"`
	Symbols           *bool `json:"symbols,omitempty"`
	Urls              *bool `json:"urls,omitempty"`
	Words             *bool `json:"words,omitempty"`
}

func UpdateAutoModSettings(req UpdateAutoModSettingsRequest) error {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
		"moderator_id":   user.UserID,
	}
	url := BuildURL(HelixBaseURL+"/automod/settings", opts)

	_, err := ExecuteJSONRequest[map[string]any, UpdateAutoModSettingsRequest]("PATCH", url, req, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] UpdateAutoModSettings: broadcasterID=%v", user.UserID)
		return err
	}

	return nil
}

func GetBannedUsers(userIDs []string, req *PaginationRequest) (*PaginationData[BannedUser], error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
	}
	for _, id := range userIDs {
		opts["user_id"] = id
	}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := BuildURL(HelixBaseURL+"/moderation/banned", opts)

	result, err := ExecuteRequest[PaginationData[BannedUser]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetBannedUsers: broadcasterID=%v", user.UserID)
		return nil, err
	}

	quantity := 0
	if req != nil {
		quantity = req.Quantity
	}
	result.GetNext = func() *PaginationData[BannedUser] {
		r, _ := GetBannedUsers(userIDs, &PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: quantity,
		})
		return r
	}

	return result, nil
}

func UnbanUser(broadcasterID, userID string) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	opts := map[string]any{
		"broadcaster_id": broadcasterID,
		"moderator_id":   user.UserID,
		"user_id":        userID,
	}
	url := BuildURL(HelixBaseURL+"/moderation/banned", opts)

	_, err := ExecuteRequest[map[string]any]("DELETE", url, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] UnbanUser: broadcasterID=%v, userID=%v", broadcasterID, userID)
		return err
	}

	return nil
}

func BanUser(userId string, duration int32, reason string) (string, error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
		"moderator_id":   user.UserID,
	}
	url := BuildURL(HelixBaseURL+"/moderation/banned", opts)

	body := map[string]any{
		"user_id": userId,
		"reason":  reason,
	}
	if duration > 0 {
		body["duration"] = duration
	}

	_, err := ExecuteJSONRequest[map[string]any, map[string]any]("POST", url, body, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] BanUser: userId=%v, duration=%v, reason=%v", userId, duration, reason)
		return "", err
	}

	return "", nil
}

func GetBlockedTerms(req *PaginationRequest) (*PaginationData[BlockedTerm], error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
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

	url := BuildURL(HelixBaseURL+"/moderation/blocked_terms", opts)

	result, err := ExecuteRequest[PaginationData[BlockedTerm]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetBlockedTerms: broadcasterID=%v", user.UserID)
		return nil, err
	}

	quantity := 0
	if req != nil {
		quantity = req.Quantity
	}
	result.GetNext = func() *PaginationData[BlockedTerm] {
		r, _ := GetBlockedTerms(&PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: quantity,
		})
		return r
	}

	return result, nil
}

func AddBlockedTerm(text string, duration int) (*BlockedTerm, error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
		"moderator_id":   user.UserID,
	}
	url := BuildURL(HelixBaseURL+"/moderation/blocked_terms", opts)

	body := map[string]any{
		"text": text,
	}
	if duration > 0 {
		body["duration_seconds"] = duration
	}

	type AddBlockedTermResponse struct {
		Data []BlockedTerm `json:"data"`
	}

	result, err := ExecuteJSONRequest[AddBlockedTermResponse, map[string]any]("POST", url, body, 201)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] AddBlockedTerm: broadcasterID=%v, text=%v", user.UserID, text)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func RemoveBlockedTerm(termID string) error {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
		"moderator_id":   user.UserID,
		"id":             termID,
	}
	url := BuildURL(HelixBaseURL+"/moderation/blocked_terms", opts)

	_, err := ExecuteRequest[map[string]any]("DELETE", url, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] RemoveBlockedTerm: broadcasterID=%v, termID=%v", user.UserID, termID)
		return err
	}

	return nil
}

func GetModeratedChannels(userID string) ([]string, error) {
	opts := map[string]any{
		"user_id": userID,
	}
	url := BuildURL(HelixBaseURL+"/moderation/channels", opts)

	type ChannelResponse struct {
		BroadcasterID string `json:"broadcaster_id"`
	}

	type GetModeratedChannelsResponse struct {
		Data []ChannelResponse `json:"data"`
	}

	result, err := ExecuteRequest[GetModeratedChannelsResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetModeratedChannels: userID=%v", userID)
		return nil, err
	}

	channels := make([]string, len(result.Data))
	for i, d := range result.Data {
		channels[i] = d.BroadcasterID
	}

	return channels, nil
}

func GetModerators(userIDs []string, req *PaginationRequest) (*PaginationData[Moderator], error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
	}
	for _, id := range userIDs {
		opts["user_id"] = id
	}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := BuildURL(HelixBaseURL+"/moderation/moderators", opts)

	result, err := ExecuteRequest[PaginationData[Moderator]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetModerators: broadcasterID=%v", user.UserID)
		return nil, err
	}

	quantity := 0
	if req != nil {
		quantity = req.Quantity
	}
	result.GetNext = func() *PaginationData[Moderator] {
		r, _ := GetModerators(userIDs, &PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: quantity,
		})
		return r
	}

	return result, nil
}

func AddChannelModerator(broadcasterID, userID string) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	opts := map[string]any{
		"broadcaster_id": broadcasterID,
		"user_id":        userID,
	}
	url := BuildURL(HelixBaseURL+"/moderation/moderators", opts)

	_, err := ExecuteRequest[map[string]any]("POST", url, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] AddChannelModerator: broadcasterID=%v, userID=%v", broadcasterID, userID)
		return err
	}

	return nil
}

func RemoveChannelModerator(broadcasterID, userID string) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	opts := map[string]any{
		"broadcaster_id": broadcasterID,
		"user_id":        userID,
	}
	url := BuildURL(HelixBaseURL+"/moderation/moderators", opts)

	_, err := ExecuteRequest[map[string]any]("DELETE", url, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] RemoveChannelModerator: broadcasterID=%v, userID=%v", broadcasterID, userID)
		return err
	}

	return nil
}

func GetVIPs(userIDs []string, req *PaginationRequest) (*PaginationData[VIP], error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
	}
	for _, id := range userIDs {
		opts["user_id"] = id
	}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := BuildURL(HelixBaseURL+"/moderation/vips", opts)

	result, err := ExecuteRequest[PaginationData[VIP]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetVIPs: broadcasterID=%v", user.UserID)
		return nil, err
	}

	quantity := 0
	if req != nil {
		quantity = req.Quantity
	}
	result.GetNext = func() *PaginationData[VIP] {
		r, _ := GetVIPs(userIDs, &PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: quantity,
		})
		return r
	}

	return result, nil
}

func AddChannelVIP(broadcasterID, userID string) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	opts := map[string]any{
		"broadcaster_id": broadcasterID,
		"user_id":        userID,
	}
	url := BuildURL(HelixBaseURL+"/moderation/vips", opts)

	_, err := ExecuteRequest[map[string]any]("POST", url, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] AddChannelVIP: broadcasterID=%v, userID=%v", broadcasterID, userID)
		return err
	}

	return nil
}

func RemoveChannelVIP(broadcasterID, userID string) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	opts := map[string]any{
		"broadcaster_id": broadcasterID,
		"user_id":        userID,
	}
	url := BuildURL(HelixBaseURL+"/moderation/vips", opts)

	_, err := ExecuteRequest[map[string]any]("DELETE", url, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] RemoveChannelVIP: broadcasterID=%v, userID=%v", broadcasterID, userID)
		return err
	}

	return nil
}

func GetShieldModeStatus() (*ShieldModeStatus, error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
		"moderator_id":   user.UserID,
	}
	url := BuildURL(HelixBaseURL+"/moderation/shield_mode", opts)

	result, err := ExecuteRequest[GetShieldModeStatusResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetShieldModeStatus: broadcasterID=%v", user.UserID)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func UpdateShieldModeStatus(isActive bool) error {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
		"moderator_id":   user.UserID,
	}
	url := BuildURL(HelixBaseURL+"/moderation/shield_mode", opts)

	body := map[string]any{
		"is_active": isActive,
	}

	_, err := ExecuteJSONRequest[map[string]any, map[string]any]("PUT", url, body, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] UpdateShieldModeStatus: broadcasterID=%v, isActive=%v", user.UserID, isActive)
		return err
	}

	return nil
}

func WarnChatUser(userID, reason string) error {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
		"moderator_id":   user.UserID,
	}
	url := BuildURL(HelixBaseURL+"/moderation/warns", opts)

	body := map[string]any{
		"user_id": userID,
		"reason":  reason,
	}

	_, err := ExecuteJSONRequest[map[string]any, map[string]any]("POST", url, body, 201)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] WarnChatUser: broadcasterID=%v, userID=%v", user.UserID, userID)
		return err
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

	opts := map[string]any{
		"broadcaster_id": broadcasterID,
		"moderator_id":   moderatorID,
	}
	url := BuildURL(HelixBaseURL+"/moderation/suspicious", opts)

	body := map[string]any{
		"user_id": userID,
	}

	_, err := ExecuteJSONRequest[map[string]any, map[string]any]("POST", url, body, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] AddSuspiciousStatusToChatUser: broadcasterID=%v, userID=%v", broadcasterID, userID)
		return err
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

	opts := map[string]any{
		"broadcaster_id": broadcasterID,
		"moderator_id":   moderatorID,
		"user_id":        userID,
	}
	url := BuildURL(HelixBaseURL+"/moderation/suspicious", opts)

	_, err := ExecuteRequest[map[string]any]("DELETE", url, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] RemoveSuspiciousStatusFromChatUser: broadcasterID=%v, userID=%v", broadcasterID, userID)
		return err
	}

	return nil
}

func GetUnbanRequests(userID, status string, req *PaginationRequest) (*PaginationData[UnbanRequest], error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
		"user_id":        userID,
		"status":         status,
	}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := BuildURL(HelixBaseURL+"/moderation/unban_requests", opts)

	result, err := ExecuteRequest[PaginationData[UnbanRequest]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetUnbanRequests: broadcasterID=%v", user.UserID)
		return nil, err
	}

	quantity := 0
	if req != nil {
		quantity = req.Quantity
	}
	result.GetNext = func() *PaginationData[UnbanRequest] {
		r, _ := GetUnbanRequests(userID, status, &PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: quantity,
		})
		return r
	}

	return result, nil
}

func ResolveUnbanRequest(requestID, action, resolutionText string) error {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
		"id":             requestID,
	}
	url := BuildURL(HelixBaseURL+"/moderation/unban_requests", opts)

	body := map[string]any{
		"status":          action,
		"resolution_text": resolutionText,
	}

	_, err := ExecuteJSONRequest[map[string]any, map[string]any]("PATCH", url, body, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] ResolveUnbanRequest: broadcasterID=%v, requestID=%v, action=%v", user.UserID, requestID, action)
		return err
	}

	return nil
}
