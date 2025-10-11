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

type UserResponse struct {
	Data []TwitchUserData `json:"data"`
}

type TwitchUserData struct {
	ID    string `json:"id"`
	Login string `json:"login"`
}

type StreamData struct {
	GameID string `json:"game_id"`
	Title  string `json:"title"`
}

type GameData struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	BoxArt string `json:"box_art_url"`
}

// getUserID retorna o ID numérico de qualquer usuário
func GetUserData(login string) (TwitchUserData, error) {
	urlAPI := fmt.Sprintf("https://api.twitch.tv/helix/users?login=%s", login)
	req, _ := http.NewRequest("GET", urlAPI, nil)
	req.Header.Set("Authorization", "Bearer "+Token)
	req.Header.Set("Client-ID", globals.GetConfig().TwitchClientID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return TwitchUserData{}, err
	}
	defer resp.Body.Close()

	var u UserResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &u)
	if len(u.Data) == 0 {
		return TwitchUserData{}, fmt.Errorf("usuário não encontrado")
	}
	return u.Data[0], nil
}

func DeleteMessage(msgID string) error {
	urlAPI := fmt.Sprintf("https://api.twitch.tv/helix/moderation/chat?broadcaster_id=%s&moderator_id=%s&message_id=%s", UserID, UserID, msgID)
	req, _ := http.NewRequest("DELETE", urlAPI, nil)
	req.Header.Set("Authorization", "Bearer "+Token)
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

func GetListOfGames(query string) ([]GameData, error) {
	urlAPI := fmt.Sprintf("%s?query=%s", urlAPIGames, query)
	req, _ := http.NewRequest("GET", urlAPI, nil)
	req.Header.Set("Authorization", "Bearer "+Token)
	req.Header.Set("Client-ID", globals.GetConfig().TwitchClientID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		helpers.Logf(helpers.Red, "[TWITCH FETCH] Erro ao buscar lista de jogos: %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		helpers.Logf(helpers.Red, "[TWITCH FETCH] Erro ao buscar lista de jogos: (%d) %s", resp.StatusCode, body)
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

func UpdateStreamData(sd StreamData) error {
	jsonData, _ := json.Marshal(sd)
	urlAPI := fmt.Sprintf("%s?broadcaster_id=%s", urlAPIChannel, UserID)
	req, _ := http.NewRequest("PATCH", urlAPI, bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+Token)
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
