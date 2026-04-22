package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var urlAPIGames = "https://api.twitch.tv/helix/search/categories"
var urlAPIChannel = "https://api.twitch.tv/helix/channels"
var urlAPIBadges = "https://api.twitch.tv/helix/chat/badges"
var urlAPIEventSub = "https://api.twitch.tv/helix/eventsub/subscriptions"
var urlAPIFollowers = "https://api.twitch.tv/helix/channels/followers"

type UserResponse struct {
	Data []TwitchUserData `json:"data"`
}

type TwitchUserData struct {
	ID                     string `json:"id"`
	Login                  string `json:"login"`
	DisplayName            string `json:"display_name"`
	Type                   string `json:"type"`
	BroadcasterType        string `json:"broadcaster_type"`
	Description            string `json:"description"`
	ProfileImageURL        string `json:"profile_image_url"`
	ProfileOfflineImageURL string `json:"offline_image_url"`
	ViewCount              int    `json:"view_count"`
	Email                  string `json:"email"`
	CreatedAt              string `json:"created_at"`
}

type TwitchViewerData struct {
	UserId     string `json:"user_id"`
	UserName   string `json:"user_name"`
	UserLogin  string `json:"user_login"`
	FollowedAt string `json:"followed_at"`
}

type StreamData struct {
	BroadcasterId               string   `json:"broadcaster_id"`
	BroadcasterLogin            string   `json:"broadcaster_login"`
	BroadcasterName             string   `json:"broadcaster_name"`
	BroadcasterLanguage         string   `json:"broadcaster_language"`
	GameID                      string   `json:"game_id"`
	GameName                    string   `json:"game_name"`
	Title                       string   `json:"title"`
	Delay                       int      `json:"delay"`
	Tags                        []string `json:"tags"`
	ContentClassificationLabels []string `json:"content_classification_labels"`
	IsBrandedContent            bool     `json:"is_branded_content"`
}

type GameData struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	BoxArt string `json:"box_art_url"`
}

func ValidateAccessToken(accessToken string) (*struct {
	ClientId  string   `json:"client_id"`
	Login     string   `json:"login"`
	Scopes    []string `json:"scopes"`
	UserId    string   `json:"user_id"`
	ExpiresIn int      `json:"expires_in"`
	Status    int      `json:"status"`
	Message   string   `json:"message"`
}, error) {
	req, err := http.NewRequest("GET", "https://id.twitch.tv/oauth2/validate", nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] ValidateAccessToken http.NewRequest failed: %v", err)
		return nil, err
	}
	req.Header.Set("Authorization", "OAuth "+accessToken)
	req.Header.Set("Client-ID", globals.GetConfig().TwitchClientID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var u struct {
		ClientId  string   `json:"client_id"`
		Login     string   `json:"login"`
		Scopes    []string `json:"scopes"`
		UserId    string   `json:"user_id"`
		ExpiresIn int      `json:"expires_in"`
		Status    int      `json:"status"`
		Message   string   `json:"message"`
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] ValidateAccessToken io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &u); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] ValidateAccessToken json.Unmarshal failed: %v", err)
		return nil, err
	}
	return &u, nil
}

func GetStreamerData() (*TwitchUserData, error) {
	req, err := http.NewRequest("GET", "https://api.twitch.tv/helix/users", nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetStreamerData http.NewRequest failed: %v", err)
		return nil, err
	}
	userResp, err := DoRequest(req)
	if err != nil {
		return nil, err
	}
	defer userResp.Body.Close()
	dataUser, err := io.ReadAll(userResp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetStreamerData io.ReadAll failed: %v", err)
		return nil, err
	}

	var u struct {
		Data []TwitchUserData `json:"data"`
	}
	if err := json.Unmarshal(dataUser, &u); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetStreamerData json.Unmarshal failed: %v", err)
		return nil, err
	}

	if len(u.Data) == 0 {
		return nil, fmt.Errorf("GetUserDataByToken: no user returned, check token and scopes")
	}

	return &u.Data[0], nil
}

func GetUserData(login string) (*TwitchUserData, error) {
	urlAPI := fmt.Sprintf("https://api.twitch.tv/helix/users?login=%s", login)
	req, err := http.NewRequest("GET", urlAPI, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetUserData http.NewRequest failed: %v", err)
		return nil, err
	}
	resp, err := DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetUserData: login=%v", login)
		return nil, err
	}
	defer resp.Body.Close()

	var u UserResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetUserData io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &u); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetUserData json.Unmarshal failed: %v", err)
		return nil, err
	}
	if len(u.Data) == 0 {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetUserData: login=%v", login)
		return nil, fmt.Errorf("GetUserData(%s): user not found", login)
	}
	return &u.Data[0], nil
}

func GetUserDataById(id string) (*TwitchUserData, error) {
	urlAPI := fmt.Sprintf("https://api.twitch.tv/helix/users?id=%s", id)
	req, err := http.NewRequest("GET", urlAPI, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetUserDataById http.NewRequest failed: %v", err)
		return nil, err
	}
	resp, err := DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetUserDataById: id=%v", id)
		return nil, err
	}
	defer resp.Body.Close()

	var u UserResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetUserDataById io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &u); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetUserDataById json.Unmarshal failed: %v", err)
		return nil, err
	}
	if len(u.Data) == 0 {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetUserDataById: id=%v", id)
		return nil, fmt.Errorf("GetUserDataById(%s): user not found", id)
	}
	return &u.Data[0], nil
}

func GetFollowersData(broadcaster_id, userId string) ([]TwitchViewerData, error) {
	url := urlAPIFollowers
	if broadcaster_id == "" {
		broadcaster_id = globals.GetState().GetTwitchUser().UserID
	}
	url = fmt.Sprintf("%s?broadcaster_id=%s", url, broadcaster_id)
	if userId != "" {
		url = fmt.Sprintf("%s&user_id=%s", url, userId)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetFollowersData http.NewRequest failed: %v", err)
		return []TwitchViewerData{}, err
	}
	resp, err := DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetFollowersData: broadcaster_id=%v, userId=%v", broadcaster_id, userId)
		return []TwitchViewerData{}, err
	}
	defer resp.Body.Close()

	var u struct {
		Data []TwitchViewerData `json:"data"`
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetFollowersData io.ReadAll failed: %v", err)
		return []TwitchViewerData{}, err
	}
	if err := json.Unmarshal(body, &u); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetFollowersData json.Unmarshal failed: %v", err)
		return []TwitchViewerData{}, err
	}
	return u.Data, nil
}

func DeleteMessage(msgID string) error {
	user := globals.GetState().TwitchUser
	urlAPI := fmt.Sprintf("https://api.twitch.tv/helix/moderation/chat?broadcaster_id=%s&moderator_id=%s&message_id=%s", user.UserID, user.UserID, msgID)
	req, err := http.NewRequest("DELETE", urlAPI, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] DeleteMessage http.NewRequest failed: %v", err)
		return err
	}
	resp, err := DoRequest(req)
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
	urlAPI := fmt.Sprintf("https://api.twitch.tv/helix/moderation/bans?broadcaster_id=%s&moderator_id=%s", user.UserID, user.UserID)
	req, err := http.NewRequest("POST", urlAPI, bytes.NewBuffer(data))
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] BanUser http.NewRequest failed: %v", err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := DoRequest(req)
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

func GetListOfGames(query string) ([]GameData, error) {
	urlAPI := fmt.Sprintf("%s?query=%s", urlAPIGames, query)
	req, err := http.NewRequest("GET", urlAPI, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetListOfGames http.NewRequest failed: %v", err)
		return nil, err
	}
	resp, err := DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetListOfGames: query=%v", query)
		helpers.Logf(helpers.ERROR, "[TWITCH FETCH] Error fetching game list: %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetListOfGames io.ReadAll failed: %v", err)
			return nil, err
		}
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetListOfGames: query=%v", query)
		helpers.Logf(helpers.ERROR, "[TWITCH FETCH] Error fetching game list: (%d) %s", resp.StatusCode, body)
		return nil, fmt.Errorf("GetListOfGames(%s): failed to fetch game list: %s", query, body)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetListOfGames io.ReadAll failed: %v", err)
		return nil, err
	}
	//helpers.Logf(helpers.Twitch, "[TWITCH FETCH] GetListOfGames: %s", body)
	var reqData struct {
		Data []GameData `json:"data"`
	}
	if err := json.Unmarshal(body, &reqData); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetListOfGames json.Unmarshal failed: %v", err)
		return nil, err
	}
	return reqData.Data, nil
}

func GetChannelStreamData(id string) (*StreamData, error) {
	urlAPI := fmt.Sprintf("%s?broadcaster_id=%s", urlAPIChannel, id)
	req, err := http.NewRequest("GET", urlAPI, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetChannelStreamData http.NewRequest failed: %v", err)
		return nil, err
	}
	resp, err := DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetChannelStreamData: id=%v", id)
		helpers.Logf(helpers.ERROR, "[TWITCH] GetStreamData: %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	var u struct {
		Data []StreamData `json:"data"`
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetChannelStreamData io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &u); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetChannelStreamData json.Unmarshal failed: %v", err)
		return nil, err
	}
	if len(u.Data) == 0 {
		helpers.Logf(helpers.ERROR, "[TWITCH] No stream data on GetStreamData")
		return nil, nil
	}
	return &u.Data[0], nil
}

func UpdateChannelStreamData(sd *StreamData) error {
	if sd == nil {
		return fmt.Errorf("UpdateChannelStreamData: nil stream data")
	}
	user := globals.GetState().GetTwitchUser()
	jsonData, err := json.Marshal(sd)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateChannelStreamData json.Marshal failed: %v", err)
		return err
	}
	urlAPI := fmt.Sprintf("%s?broadcaster_id=%s", urlAPIChannel, user.UserID)
	req, err := http.NewRequest("PATCH", urlAPI, bytes.NewBuffer(jsonData))
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateChannelStreamData http.NewRequest failed: %v", err)
		return err
	}
	resp, err := DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] UpdateChannelStreamData: sd=%v", sd)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] UpdateChannelStreamData io.ReadAll failed: %v", err)
			return err
		}
		helpers.Logf(helpers.DEBUG, "[TWITCH] UpdateChannelStreamData: sd=%v", sd)
		return fmt.Errorf("UpdateChannelStreamData(%+v): failed to update: %s", sd, body)
	}
	return nil
}

func GetBadges(broadcasterId ...string) (*map[string]any, error) {
	urlAPI := urlAPIBadges
	if len(broadcasterId) > 0 {
		urlAPI = fmt.Sprintf(urlAPI+"?broadcaster_id=%s", broadcasterId[0])
	} else {
		urlAPI = urlAPI + "/global"
	}

	req, err := http.NewRequest("GET", urlAPI, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetBadges http.NewRequest failed: %v", err)
		return nil, err
	}
	resp, err := DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetBadges: broadcasterId=%v", broadcasterId)
		helpers.Logf(helpers.ERROR, "[TWITCH FETCH] Error fetching badges list: %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetBadges io.ReadAll failed: %v", err)
			return nil, err
		}
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetBadges: broadcasterId=%v", broadcasterId)
		helpers.Logf(helpers.ERROR, "[TWITCH FETCH] Error fetching badges: (%d) %s", resp.StatusCode, body)
		return nil, fmt.Errorf("GetBadges(%v): failed to fetch badges: %s", broadcasterId, body)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetBadges io.ReadAll failed: %v", err)
		return nil, err
	}
	//helpers.Logf(helpers.Twitch, "[TWITCH FETCH] GetBadges: %s", body)
	var reqData struct {
		Data []struct {
			SetId    string `json:"set_id"`
			Versions []struct {
				Id          int    `json:"id"`
				ImgUrl1x    string `json:"image_url_1x"`
				ImgUrl2x    string `json:"image_url_2x"`
				ImgUrl4x    string `json:"image_url_4x"`
				Title       string `json:"title"`
				Description string `json:"description"`
				ClickAction string `json:"click_action"`
				ClickUrl    string `json:"click_url"`
			} `json:"versions"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &reqData); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetBadges json.Unmarshal failed: %v", err)
		return nil, err
	}

	d := make(map[string]any)
	for _, v := range reqData.Data {
		d[v.SetId] = v.Versions
	}

	return &d, nil
}

func GetEventSubscriptions() (*EventSubData, error) {
	req, err := http.NewRequest("GET", urlAPIEventSub, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetEventSubscriptions http.NewRequest failed: %v", err)
		return nil, err
	}
	resp, err := DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetEventSubscriptions: no params")
		helpers.Logf(helpers.ERROR, "[TWITCH FETCH] Error fetching event subscriptions list: %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetEventSubscriptions io.ReadAll failed: %v", err)
			return nil, err
		}
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetEventSubscriptions: no params")
		helpers.Logf(helpers.ERROR, "[TWITCH FETCH] Error fetching event subscriptions: (%d) %s", resp.StatusCode, body)
		return nil, fmt.Errorf("GetEventSubscriptions: failed to fetch subscriptions: %s", body)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetEventSubscriptions io.ReadAll failed: %v", err)
		return nil, err
	}
	//helpers.Logf(helpers.Twitch, "[TWITCH FETCH] GetBadges: %s", body)
	var reqData EventSubData
	if err := json.Unmarshal(body, &reqData); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetEventSubscriptions json.Unmarshal failed: %v", err)
		return nil, err
	}

	return &reqData, nil
}

func DeleteEventSubscriptions(id string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s?id=%s", urlAPIEventSub, id), nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] DeleteEventSubscriptions http.NewRequest failed: %v", err)
		return err
	}
	resp, err := DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] DeleteEventSubscriptions: id=%v", id)
		helpers.Logf(helpers.ERROR, "[TWITCH FETCH] Error deleting event subscription: %s", err.Error())
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] DeleteEventSubscriptions io.ReadAll failed: %v", err)
			return err
		}
		helpers.Logf(helpers.DEBUG, "[TWITCH] DeleteEventSubscriptions: id=%v", id)
		helpers.Logf(helpers.ERROR, "[TWITCH FETCH] Error deleting event subscriptions: (%d) %s", resp.StatusCode, body)
		return fmt.Errorf("DeleteEventSubscriptions(%s): failed to delete: %s", id, body)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] DeleteEventSubscriptions io.ReadAll failed: %v", err)
		return err
	}
	//helpers.Logf(helpers.Twitch, "[TWITCH FETCH] GetBadges: %s", body)
	var reqData EventSubData
	if err := json.Unmarshal(body, &reqData); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] DeleteEventSubscriptions json.Unmarshal failed: %v", err)
		return err
	}

return nil
}

func GetUserChatColor(id string) (*struct {
	UserId    string `json:"user_id"`
	UserName  string `json:"user_name"`
	UserLogin string `json:"user_login"`
	Color     string `json:"color"`
}, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.twitch.tv/helix/chat/color?user_id=%s", id), nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetUserChatColor http.NewRequest failed: %v", err)
		return nil, err
	}
	resp, err := DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetUserChatColor: id=%v", id)
		helpers.Logf(helpers.ERROR, "[TWITCH FETCH] Error fetching user color: %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetUserChatColor io.ReadAll failed: %v", err)
			return nil, err
		}
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetUserChatColor: id=%v", id)
		helpers.Logf(helpers.ERROR, "[TWITCH FETCH] Error fetching user color: (%d) %s", resp.StatusCode, body)
		return nil, fmt.Errorf("GetUserChatColor(%s): failed to fetch user color: %s", id, body)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetUserChatColor io.ReadAll failed: %v", err)
		return nil, err
	}
	var d struct {
		Data []struct {
			UserId    string `json:"user_id"`
			UserName  string `json:"user_name"`
			UserLogin string `json:"user_login"`
			Color     string `json:"color"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &d); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetUserChatColor json.Unmarshal failed: %v", err)
		return nil, err
	}

	if len(d.Data) == 0 {
		return nil, fmt.Errorf("GetUserChatColor(%s): no items found", id)
	}

	return &d.Data[0], nil
}

func GetStreamData(id string) (*globals.TwitchStreamData, error) {
	urlAPI := fmt.Sprintf("https://api.twitch.tv/helix/streams?user_id=%s", id)
	req, err := http.NewRequest("GET", urlAPI, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetStreamData http.NewRequest failed: %v", err)
		return nil, err
	}
	resp, err := DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetStreamData: id=%v", id)
		helpers.Logf(helpers.ERROR, "[TWITCH] GetStreamData: %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	var u struct {
		Data []globals.TwitchStreamData `json:"data"`
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetStreamData io.ReadAll failed: %v", err)
		return nil, err
	}
if err := json.Unmarshal(body, &u); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetStreamData json.Unmarshal failed: %v", err)
		return nil, err
	}
	if len(u.Data) == 0 {
		helpers.Logf(helpers.ERROR, "[TWITCH] No stream data on GetStreamData")
		return nil, nil
	}
	return &u.Data[0], nil
}

func UpdateAutomod(userId, msgId, action string) (string, error) {
	d := map[string]any{
		"user_id": globals.GetState().GetTwitchUser().UserID,
		"msg_id":  msgId,
		"action":  action,
	}
	data, err := json.Marshal(d)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateAutomod json.Marshal failed: %v", err)
		return "", err
	}
	req, err := http.NewRequest("POST", "https://api.twitch.tv/helix/moderation/automod/message", bytes.NewBuffer(data))
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateAutomod http.NewRequest failed: %v", err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] UpdateAutomod: userId=%v, msgId=%v, action=%v", userId, msgId, action)
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] UpdateAutomod io.ReadAll failed: %v", err)
			return "", err
		}
		helpers.Logf(helpers.DEBUG, "[TWITCH] UpdateAutomod: userId=%v, msgId=%v, action=%v", userId, msgId, action)
		return "", fmt.Errorf("UpdateAutomod(%s, %s, %s): failed to update automod: %s", userId, msgId, action, body)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateAutomod io.ReadAll failed: %v", err)
		return "", err
	}
	return string(body), nil
}
