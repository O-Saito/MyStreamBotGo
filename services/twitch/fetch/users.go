package twitch

import (
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var urlAPIUsers = "https://api.twitch.tv/helix/users"
var urlAPIUserBlock = "https://api.twitch.tv/helix/users/block"
var urlAPIUserExtensions = "https://api.twitch.tv/helix/users/extensions"

type User struct {
	ID            string `json:"id"`
	Login         string `json:"login"`
	DisplayName   string `json:"display_name"`
	Type          string `json:"type"`
	BroadcasterType string `json:"broadcaster_type"`
	Description   string `json:"description"`
	ProfileImageURL string `json:"profile_image_url"`
	OfflineImageURL string `json:"offline_image_url"`
	ViewCount     int    `json:"view_count"`
	Email         string `json:"email"`
	CreatedAt     string `json:"created_at"`
}

type UserExtension struct {
	Active bool `json:"active"`
	ID     string `json:"id"`
	Name   string `json:"name"`
	Version string `json:"version"`
}

type UserActiveExtension struct {
	Active        bool          `json:"active"`
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Version       string        `json:"version"`
	ConfigureURL  string        `json:"configure_url"`
	ModuleURL     string        `json:"module_url"`
	Panel         UserExtensionSettings `json:"panel"`
	Component     UserExtensionSettings `json:"component"`
	VideoOverlay  UserExtensionSettings `json:"video_overlay"`
	Chat          UserExtensionSettings `json:"chat"`
}

type UserExtensionSettings struct {
	Active bool `json:"active"`
	Height *int `json:"height,omitempty"`
	PositionX *int `json:"position_x,omitempty"`
	PositionY *int `json:"position_y,omitempty"`
}

type UserBlock struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
}

type UserAuthorization struct {
	ClientID      string   `json:"client_id"`
	UserID        string   `json:"user_id"`
	Scopes        []string `json:"scopes"`
	ExpiresAt     string   `json:"expires_at"`
}

type GetUsersResponse struct {
	Data []User `json:"data"`
}

type GetUserBlockListResponse struct {
	Data []UserBlock `json:"data"`
}

type GetUserExtensionsResponse struct {
	Data []UserExtension `json:"data"`
}

type GetUserActiveExtensionsResponse struct {
	Data map[string]UserActiveExtension `json:"data"`
}

type GetUserAuthorizationResponse struct {
	Data []UserAuthorization `json:"data"`
}

func UpdateUser(description string) (*User, error) {
	data := map[string]any{
		"description": description,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateUser json.Marshal failed: %v", err)
		return nil, err
	}

	req, err := http.NewRequest("PUT", urlAPIUsers, bytes.NewBuffer(jsonData))
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateUser http.NewRequest failed: %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] UpdateUser: description=%v", description)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] UpdateUser io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("UpdateUser: failed: %s", body)
	}

	var result GetUsersResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateUser io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateUser json.Unmarshal failed: %v", err)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func GetAuthorizationByUser(userID string) (*UserAuthorization, error) {
	url := fmt.Sprintf("%s?id=%s", urlAPIUsers+"/authorized", userID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetAuthorizationByUser http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetAuthorizationByUser: userID=%v", userID)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetAuthorizationByUser io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("GetAuthorizationByUser: failed: %s", body)
	}

	var result GetUserAuthorizationResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetAuthorizationByUser io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetAuthorizationByUser json.Unmarshal failed: %v", err)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func GetUserBlockList(targetUserID string) ([]UserBlock, error) {
	url := fmt.Sprintf("%s?target_user_id=%s", urlAPIUserBlock, targetUserID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetUserBlockList http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetUserBlockList: targetUserID=%v", targetUserID)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetUserBlockList io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("GetUserBlockList: failed: %s", body)
	}

	var result GetUserBlockListResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetUserBlockList io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetUserBlockList json.Unmarshal failed: %v", err)
		return nil, err
	}

	return result.Data, nil
}

func BlockUser(targetUserID string) error {
	url := fmt.Sprintf("%s?target_user_id=%s", urlAPIUserBlock, targetUserID)
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] BlockUser http.NewRequest failed: %v", err)
		return err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] BlockUser: targetUserID=%v", targetUserID)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] BlockUser io.ReadAll failed: %v", err)
			return err
		}
		return fmt.Errorf("BlockUser: failed: %s", body)
	}

	return nil
}

func UnblockUser(targetUserID string) error {
	url := fmt.Sprintf("%s?target_user_id=%s", urlAPIUserBlock, targetUserID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UnblockUser http.NewRequest failed: %v", err)
		return err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] UnblockUser: targetUserID=%v", targetUserID)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] UnblockUser io.ReadAll failed: %v", err)
			return err
		}
		return fmt.Errorf("UnblockUser: failed: %s", body)
	}

	return nil
}

func GetUserExtensions() ([]UserExtension, error) {
	req, err := http.NewRequest("GET", urlAPIUserExtensions, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetUserExtensions http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetUserExtensions: no params")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetUserExtensions io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("GetUserExtensions: failed: %s", body)
	}

	var result GetUserExtensionsResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetUserExtensions io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetUserExtensions json.Unmarshal failed: %v", err)
		return nil, err
	}

	return result.Data, nil
}

func GetUserActiveExtensions() (map[string]UserActiveExtension, error) {
	req, err := http.NewRequest("GET", urlAPIUserExtensions+"?active=true", nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetUserActiveExtensions http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetUserActiveExtensions: no params")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetUserActiveExtensions io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("GetUserActiveExtensions: failed: %s", body)
	}

	var result GetUserActiveExtensionsResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetUserActiveExtensions io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetUserActiveExtensions json.Unmarshal failed: %v", err)
		return nil, err
	}

	return result.Data, nil
}

func UpdateUserExtensions(extensions map[string]UserActiveExtension) error {
	data := map[string]any{
		"data": extensions,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateUserExtensions json.Marshal failed: %v", err)
		return err
	}

	req, err := http.NewRequest("PUT", urlAPIUserExtensions, bytes.NewBuffer(jsonData))
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateUserExtensions http.NewRequest failed: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] UpdateUserExtensions: extensions count=%v", len(extensions))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] UpdateUserExtensions io.ReadAll failed: %v", err)
			return err
		}
		return fmt.Errorf("UpdateUserExtensions: failed: %s", body)
	}

	return nil
}

func GetUserData(login string) (*User, error) {
	urlAPI := fmt.Sprintf("%s?login=%s", urlAPIUsers, login)
	req, err := http.NewRequest("GET", urlAPI, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetUserData http.NewRequest failed: %v", err)
		return nil, err
	}
	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetUserData: login=%v", login)
		return nil, err
	}
	defer resp.Body.Close()

	var u struct {
		Data []User `json:"data"`
	}
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

func GetUserDataById(id string) (*User, error) {
	urlAPI := fmt.Sprintf("%s?id=%s", urlAPIUsers, id)
	req, err := http.NewRequest("GET", urlAPI, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetUserDataById http.NewRequest failed: %v", err)
		return nil, err
	}
	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetUserDataById: id=%v", id)
		return nil, err
	}
	defer resp.Body.Close()

	var u struct {
		Data []User `json:"data"`
	}
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