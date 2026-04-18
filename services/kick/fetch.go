package kick

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
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
	resp, err := DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[KICK] GetUser: userId=%v", userId)
		return UserData{}, err
	}
	defer resp.Body.Close()

	var u UserDataResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[KICK] GetUser: userId=%v", userId)
		helpers.Logf(helpers.ERROR, "[KICK] GetUser io.ReadAll failed: %v", err)
		return UserData{}, err
	}
	if err := json.Unmarshal(body, &u); err != nil {
		helpers.Logf(helpers.DEBUG, "[KICK] GetUser: userId=%v", userId)
		helpers.Logf(helpers.ERROR, "[KICK] GetUser json.Unmarshal failed: %v", err)
		return UserData{}, err
	}
	if len(u.Data) == 0 {
		return UserData{}, fmt.Errorf("channel not found")
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
	resp, err := DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[KICK] GetChannel: streamerId=%v, slug=%v", streamerId, slug)
		return ChannelData{}, err
	}
	defer resp.Body.Close()

	var u ChannelDataResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[KICK] GetChannel: streamerId=%v, slug=%v", streamerId, slug)
		helpers.Logf(helpers.ERROR, "[KICK] GetChannel io.ReadAll failed: %v", err)
		return ChannelData{}, err
	}
	if err := json.Unmarshal(body, &u); err != nil {
		helpers.Logf(helpers.DEBUG, "[KICK] GetChannel: streamerId=%v, slug=%v", streamerId, slug)
		helpers.Logf(helpers.ERROR, "[KICK] GetChannel json.Unmarshal failed: %v", err)
		return ChannelData{}, err
	}
	if len(u.Data) == 0 {
		return ChannelData{}, fmt.Errorf("channel not found")
	}
	return u.Data[0], nil
}

func GetChatroom(slug string) (ChatroomData, error) {
	url := fmt.Sprintf("https://api.kick.com/public/v1/channels?slug=%s", slug)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Client-Id", globals.GetConfig().KickClientID)
	resp, err := DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[KICK] GetChatroom: slug=%v", slug)
		return ChatroomData{}, err
	}
	defer resp.Body.Close()

	var u ChatroomData
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[KICK] GetChatroom: slug=%v", slug)
		helpers.Logf(helpers.ERROR, "[KICK] GetChatroom io.ReadAll failed: %v", err)
		return ChatroomData{}, err
	}
	if err := json.Unmarshal(body, &u); err != nil {
		helpers.Logf(helpers.DEBUG, "[KICK] GetChatroom: slug=%v", slug)
		helpers.Logf(helpers.ERROR, "[KICK] GetChatroom json.Unmarshal failed: %v", err)
		return ChatroomData{}, err
	}
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
	req.Header.Set("Content-Type", "application/json")
	resp, err := DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[KICK] PostMessage: msg.Text=%v", msg.Text)
		return err
	}
	defer resp.Body.Close()

	var u ChatroomData
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[KICK] PostMessage: msg.Text=%v", msg.Text)
		helpers.Logf(helpers.ERROR, "[KICK] PostMessage io.ReadAll failed: %v", err)
		return err
	}
	if err := json.Unmarshal(body, &u); err != nil {
		helpers.Logf(helpers.DEBUG, "[KICK] PostMessage: msg.Text=%v", msg.Text)
		helpers.Logf(helpers.ERROR, "[KICK] PostMessage json.Unmarshal failed: %v", err)
		return err
	}
	return nil
}
