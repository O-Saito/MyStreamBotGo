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
	req, _ := http.NewRequest("GET", "https://id.twitch.tv/oauth2/validate", nil)
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
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &u)
	return &u, nil
}

func GetStreamerData() (*TwitchUserData, error) {
	req, _ := http.NewRequest("GET", "https://api.twitch.tv/helix/users", nil)
	req.Header.Set("Authorization", "Bearer "+globals.GetState().GetTwitchUser().Token)
	req.Header.Set("Client-ID", globals.GetConfig().TwitchClientID)
	userResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer userResp.Body.Close()
	dataUser, _ := io.ReadAll(userResp.Body)

	var u struct {
		Data []TwitchUserData `json:"data"`
	}
	json.Unmarshal(dataUser, &u)

	if len(u.Data) == 0 {
		return nil, fmt.Errorf("nenhum usuário retornado, verifique token e scopes")
	}

	return &u.Data[0], nil
}

func GetUserData(login string) (*TwitchUserData, error) {
	urlAPI := fmt.Sprintf("https://api.twitch.tv/helix/users?login=%s", login)
	req, _ := http.NewRequest("GET", urlAPI, nil)
	req.Header.Set("Authorization", "Bearer "+globals.GetState().GetTwitchUser().Token)
	req.Header.Set("Client-ID", globals.GetConfig().TwitchClientID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var u UserResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &u)
	if len(u.Data) == 0 {
		return nil, fmt.Errorf("usuário não encontrado")
	}
	return &u.Data[0], nil
}

func GetUserDataById(id string) (*TwitchUserData, error) {
	urlAPI := fmt.Sprintf("https://api.twitch.tv/helix/users?id=%s", id)
	req, _ := http.NewRequest("GET", urlAPI, nil)
	req.Header.Set("Authorization", "Bearer "+globals.GetState().GetTwitchUser().Token)
	req.Header.Set("Client-ID", globals.GetConfig().TwitchClientID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var u UserResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &u)
	if len(u.Data) == 0 {
		return nil, fmt.Errorf("usuário não encontrado")
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

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+globals.GetState().GetTwitchUser().Token)
	req.Header.Set("Client-ID", globals.GetConfig().TwitchClientID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []TwitchViewerData{}, err
	}
	defer resp.Body.Close()

	var u struct {
		Data []TwitchViewerData `json:"data"`
	}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &u)
	return u.Data, nil
}

func DeleteMessage(msgID string) error {
	user := globals.GetState().TwitchUser
	urlAPI := fmt.Sprintf("https://api.twitch.tv/helix/moderation/chat?broadcaster_id=%s&moderator_id=%s&message_id=%s", user.UserID, user.UserID, msgID)
	req, _ := http.NewRequest("DELETE", urlAPI, nil)
	req.Header.Set("Authorization", "Bearer "+user.Token)
	req.Header.Set("Client-ID", globals.GetConfig().TwitchClientID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("erro ao excluir mensagem: %s", body)
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

	data, _ := json.Marshal(d)
	urlAPI := fmt.Sprintf("https://api.twitch.tv/helix/moderation/bans?broadcaster_id=%s&moderator_id=%s", user.UserID, user.UserID)
	req, _ := http.NewRequest("POST", urlAPI, bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+user.Token)
	req.Header.Set("Client-ID", globals.GetConfig().TwitchClientID)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("erro ao banir usuario: %s", body)
	}
	body, _ := io.ReadAll(resp.Body)
	return string(body), nil
}

func GetListOfGames(query string) ([]GameData, error) {
	urlAPI := fmt.Sprintf("%s?query=%s", urlAPIGames, query)
	req, _ := http.NewRequest("GET", urlAPI, nil)
	req.Header.Set("Authorization", "Bearer "+globals.GetState().TwitchUser.Token)
	req.Header.Set("Client-ID", globals.GetConfig().TwitchClientID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH FETCH] Erro ao buscar lista de jogos: %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		helpers.Logf(helpers.ERROR, "[TWITCH FETCH] Erro ao buscar lista de jogos: (%d) %s", resp.StatusCode, body)
		return nil, fmt.Errorf("erro ao buscar lista de jogos: %s", body)
	}
	body, _ := io.ReadAll(resp.Body)
	//helpers.Logf(helpers.Twitch, "[TWITCH FETCH] GetListOfGames: %s", body)
	var reqData struct {
		Data []GameData `json:"data"`
	}
	_ = json.Unmarshal(body, &reqData)
	return reqData.Data, nil
}

func GetChannelStreamData(id string) (*StreamData, error) {
	urlAPI := fmt.Sprintf("%s?broadcaster_id=%s", urlAPIChannel, id)
	req, _ := http.NewRequest("GET", urlAPI, nil)
	req.Header.Set("Authorization", "Bearer "+globals.GetState().GetTwitchUser().Token)
	req.Header.Set("Client-ID", globals.GetConfig().TwitchClientID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetStreamData: %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	var u struct {
		Data []StreamData `json:"data"`
	}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &u)
	if len(u.Data) == 0 {
		helpers.Logf(helpers.ERROR, "[TWITCH] No stream data on GetStreamData")
		return nil, nil
	}
	return &u.Data[0], nil
}

func UpdateChannelStreamData(sd *StreamData) error {
	if sd == nil {
		return fmt.Errorf("sem dados para atualizar")
	}
	user := globals.GetState().GetTwitchUser()
	jsonData, _ := json.Marshal(sd)
	urlAPI := fmt.Sprintf("%s?broadcaster_id=%s", urlAPIChannel, user.UserID)
	req, _ := http.NewRequest("PATCH", urlAPI, bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+user.Token)
	req.Header.Set("Client-ID", globals.GetConfig().TwitchClientID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("erro ao excluir mensagem: %s", body)
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

	req, _ := http.NewRequest("GET", urlAPI, nil)
	req.Header.Set("Authorization", "Bearer "+globals.GetState().GetTwitchUser().Token)
	req.Header.Set("Client-ID", globals.GetConfig().TwitchClientID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH FETCH] Erro ao buscar lista de badges: %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		helpers.Logf(helpers.ERROR, "[TWITCH FETCH] Erro ao buscar lista de badges: (%d) %s", resp.StatusCode, body)
		return nil, fmt.Errorf("erro ao buscar lista de badges: %s", body)
	}
	body, _ := io.ReadAll(resp.Body)
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
	_ = json.Unmarshal(body, &reqData)

	d := make(map[string]any)
	for _, v := range reqData.Data {
		d[v.SetId] = v.Versions
	}

	return &d, nil
}

func GetEventSubscriptions() (*EventSubData, error) {
	req, _ := http.NewRequest("GET", urlAPIEventSub, nil)
	req.Header.Set("Authorization", "Bearer "+globals.GetState().GetTwitchUser().Token)
	req.Header.Set("Client-ID", globals.GetConfig().TwitchClientID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH FETCH] Erro ao buscar lista de event subscriptions: %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		helpers.Logf(helpers.ERROR, "[TWITCH FETCH] Erro ao buscar lsita de event subscriptions: (%d) %s", resp.StatusCode, body)
		return nil, fmt.Errorf("erro ao buscar lista de event subscriptions: %s", body)
	}
	body, _ := io.ReadAll(resp.Body)
	//helpers.Logf(helpers.Twitch, "[TWITCH FETCH] GetBadges: %s", body)
	var reqData EventSubData
	_ = json.Unmarshal(body, &reqData)

	return &reqData, nil
}

func DeleteEventSubscriptions(id string) error {
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s?id=%s", urlAPIEventSub, id), nil)
	req.Header.Set("Authorization", "Bearer "+globals.GetState().GetTwitchUser().Token)
	req.Header.Set("Client-ID", globals.GetConfig().TwitchClientID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH FETCH] Erro ao deletar event subscriptions (%s): %s", id, err.Error())
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		helpers.Logf(helpers.ERROR, "[TWITCH FETCH] Erro ao deletar event subscriptions: (%d) %s", resp.StatusCode, body)
		return fmt.Errorf("erro ao deletar event subscriptions: %s", body)
	}
	body, _ := io.ReadAll(resp.Body)
	//helpers.Logf(helpers.Twitch, "[TWITCH FETCH] GetBadges: %s", body)
	var reqData EventSubData
	_ = json.Unmarshal(body, &reqData)

	return nil
}

func GetUserChatColor(id string) (*struct {
	UserId    string `json:"user_id"`
	UserName  string `json:"user_name"`
	UserLogin string `json:"user_login"`
	Color     string `json:"color"`
}, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.twitch.tv/helix/chat/color?user_id=%s", id), nil)
	req.Header.Set("Authorization", "Bearer "+globals.GetState().GetTwitchUser().Token)
	req.Header.Set("Client-ID", globals.GetConfig().TwitchClientID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH FETCH] Erro ao buscar cor do usuario: %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		helpers.Logf(helpers.ERROR, "[TWITCH FETCH] Erro ao buscar cor do usuario: (%d) %s", resp.StatusCode, body)
		return nil, fmt.Errorf("erro ao buscar cor do usuario: %s", body)
	}
	body, _ := io.ReadAll(resp.Body)
	var d struct {
		Data []struct {
			UserId    string `json:"user_id"`
			UserName  string `json:"user_name"`
			UserLogin string `json:"user_login"`
			Color     string `json:"color"`
		} `json:"data"`
	}
	_ = json.Unmarshal(body, &d)

	if len(d.Data) == 0 {
		return nil, fmt.Errorf("nenhum item encontrado")
	}

	return &d.Data[0], nil
}

func GetStreamData(id string) (*globals.TwitchStreamData, error) {
	urlAPI := fmt.Sprintf("https://api.twitch.tv/helix/streams?user_id=%s", id)
	req, _ := http.NewRequest("GET", urlAPI, nil)
	req.Header.Set("Authorization", "Bearer "+globals.GetState().GetTwitchUser().Token)
	req.Header.Set("Client-ID", globals.GetConfig().TwitchClientID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetStreamData: %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	var u struct {
		Data []globals.TwitchStreamData `json:"data"`
	}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &u)
	if len(u.Data) == 0 {
		helpers.Logf(helpers.ERROR, "[TWITCH] No stream data on GetStreamData")
		return nil, nil
	}
	return &u.Data[0], nil
}

func UpdateAutomod(userId, msgId, action string) (string, error) {
	user := globals.GetState().TwitchUser
	d := map[string]any{
		"user_id": globals.GetState().GetTwitchUser().UserID,
		"msg_id":  msgId,
		"action":  action,
	}
	data, _ := json.Marshal(d)
	req, _ := http.NewRequest("POST", "https://api.twitch.tv/helix/moderation/automod/message", bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+user.Token)
	req.Header.Set("Client-ID", globals.GetConfig().TwitchClientID)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("erro ao atualizar automod: %s", body)
	}
	body, _ := io.ReadAll(resp.Body)
	return string(body), nil
}
