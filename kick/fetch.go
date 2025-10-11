package kick

import (
	"MyStreamBot/globals"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var urlAPIUsers = "https://api.kick.com/public/v1/users"
var urlAPIChannel = "https://api.kick.com/public/v1/channels"
var urlAPIChat = "https://api.kick.com/public/v1/chat"

type UserDataResponse struct {
	Data    []UserData `json:"data"`
	Message string     `json:"message"`
}

type UserData struct {
	UserId         int    `json:"user_id"`
	Email          string `json:"email"`
	Name           string `json:"name"`
	ProfilePicture string `json:"profile_picture"`
}

type ChannelDataResponse struct {
	Data    []ChannelData `json:"data"`
	Message string        `json:"message"`
}

type ChannelCategory struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Thumbnail string `json:"thumbnail"`
}

type ChannelStream struct {
	IsLive      bool   `json:"is_live"`
	IsMature    bool   `json:"is_mature"`
	Key         string `json:"key"`
	Language    string `json:"language"`
	StartTime   string `json:"start_time"`
	Thumbnail   string `json:"thumbnail"`
	Url         string `json:"url"`
	ViewerCount int    `json:"viewer_count"`
}

type ChannelData struct {
	BroadcasterUserId  int             `json:"broadcaster_user_id"`
	Slug               string          `json:"slug"`
	ChannelDescription string          `json:"channel_description"`
	BannerPicture      string          `json:"banner_picture"`
	Stream             ChannelStream   `json:"stream"`
	StreamTitle        string          `json:"stream_title"`
	Category           ChannelCategory `json:"category"`
}

type ChatroomData struct {
	ID            int    `json:"id"`
	PinnedMessage string `json:"pinned_message"`
}

func GetUser(userId string) (UserData, error) {
	url := urlAPIUsers
	if userId != "" {
		url += fmt.Sprintf("?broadcaster_user_id=%s", userId)
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return UserData{}, err
	}
	defer resp.Body.Close()

	var u UserDataResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &u)
	if len(u.Data) == 0 {
		return UserData{}, fmt.Errorf("canal não encontrado")
	}
	return u.Data[0], nil
}

func GetChannel(streamerId int, slug *string) (ChannelData, error) {
	url := urlAPIChannel
	if streamerId != 0 {
		url += fmt.Sprintf("?broadcaster_user_id=%d", streamerId)
	}
	if slug != nil {
		url += fmt.Sprintf("?slug=%s", *slug)
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ChannelData{}, err
	}
	defer resp.Body.Close()

	var u ChannelDataResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &u)
	if len(u.Data) == 0 {
		return ChannelData{}, fmt.Errorf("canal não encontrado")
	}
	return u.Data[0], nil
}

func GetChatroom(slug string) (ChatroomData, error) {
	url := fmt.Sprintf("https://api.kick.com/public/v1/channels?slug=%s", slug)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+Token)
	req.Header.Set("Client-Id", globals.GetConfig().KickClientID)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ChatroomData{}, err
	}
	defer resp.Body.Close()

	var u ChatroomData
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &u)
	return u, nil
}

func PostMessage(msg Message) error {
	url := urlAPIChat
	var data = map[string]any{
		"broadcaster_user_id": UserID,
		"content":             msg.Text,
		"type":                "user",
	}
	jsonData, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+Token)
	//req.Header.Set("Client-Id", ClientID)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var u ChatroomData
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &u)
	return nil
}
