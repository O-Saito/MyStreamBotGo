package twitch

import (
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
)

var urlAPIUsers = twitch.HelixBaseURL + "/users"
var urlAPIUserBlock = twitch.HelixBaseURL + "/users/block"
var urlAPIUserExtensions = twitch.HelixBaseURL + "/users/extensions"

type User struct {
	ID              string `json:"id"`
	Login           string `json:"login"`
	DisplayName     string `json:"display_name"`
	Type            string `json:"type"`
	BroadcasterType string `json:"broadcaster_type"`
	Description     string `json:"description"`
	ProfileImageURL string `json:"profile_image_url"`
	OfflineImageURL string `json:"offline_image_url"`
	ViewCount       int    `json:"view_count"`
	Email           string `json:"email"`
	CreatedAt       string `json:"created_at"`
}

type UserExtension struct {
	CanActivate bool     `json:"can_activate"`
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Type        []string `json:"type"`
}

type UserActiveExtension struct {
	Active       bool                  `json:"active"`
	ID           string                `json:"id"`
	Name         string                `json:"name"`
	Version      string                `json:"version"`
	ConfigureURL string                `json:"configure_url"`
	ModuleURL    string                `json:"module_url"`
	Panel        UserExtensionSettings `json:"panel"`
	Component    UserExtensionSettings `json:"component"`
	VideoOverlay UserExtensionSettings `json:"video_overlay"`
	Chat         UserExtensionSettings `json:"chat"`
}

type UserExtensionSettings struct {
	Active    bool `json:"active"`
	Height    *int `json:"height,omitempty"`
	PositionX *int `json:"position_x,omitempty"`
	PositionY *int `json:"position_y,omitempty"`
}

type UserExtensionSlotData struct {
	Active  bool   `json:"active"`
	ID      string `json:"id,omitempty"`
	Version string `json:"version,omitempty"`
	Name    string `json:"name,omitempty"`
	X       int    `json:"x,omitempty"`
	Y       int    `json:"y,omitempty"`
}

type ExtensionSlotType map[string]UserExtensionSlotData

type UserExtensionData struct {
	Panel        ExtensionSlotType `json:"panel,omitempty"`
	Overlay      ExtensionSlotType `json:"overlay,omitempty"`
	Component    ExtensionSlotType `json:"component,omitempty"`
	VideoOverlay ExtensionSlotType `json:"video_overlay,omitempty"`
	Chat         ExtensionSlotType `json:"chat,omitempty"`
}

type UserActiveExtensionsResponse struct {
	Data UserExtensionData `json:"data"`
}

type UserBlock struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
}

type UserAuthorization struct {
	ClientID  string   `json:"client_id"`
	UserID    string   `json:"user_id"`
	Scopes    []string `json:"scopes"`
	ExpiresAt string   `json:"expires_at"`
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

type GetUserAuthorizationResponse struct {
	Data []UserAuthorization `json:"data"`
}

func UpdateUser(description string) (*User, error) {
	opts := map[string]any{}
	url := twitch.BuildURL(urlAPIUsers, opts)

	body := map[string]any{"description": description}
	result, err := twitch.ExecuteJSONRequest[GetUsersResponse, map[string]any]("PUT", url, body, 200)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateUser: description=%v", description)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func GetAuthorizationByUser(userID string) (*UserAuthorization, error) {
	opts := map[string]any{
		"id": userID,
	}
	url := twitch.BuildURL(twitch.HelixBaseURL+"/users/authorized", opts)

	result, err := twitch.ExecuteRequest[GetUserAuthorizationResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetAuthorizationByUser: userID=%v", userID)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func GetUserBlockList(targetUserID string) ([]UserBlock, error) {
	opts := map[string]any{
		"target_user_id": targetUserID,
	}
	url := twitch.BuildURL(urlAPIUserBlock, opts)

	result, err := twitch.ExecuteRequest[GetUserBlockListResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetUserBlockList: targetUserID=%v", targetUserID)
		return nil, err
	}

	return result.Data, nil
}

func BlockUser(targetUserID string) error {
	opts := map[string]any{
		"target_user_id": targetUserID,
	}
	url := twitch.BuildURL(urlAPIUserBlock, opts)

	_, err := twitch.ExecuteRequest[struct{}]("PUT", url, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] BlockUser: targetUserID=%v", targetUserID)
		return err
	}

	return nil
}

func UnblockUser(targetUserID string) error {
	opts := map[string]any{
		"target_user_id": targetUserID,
	}
	url := twitch.BuildURL(urlAPIUserBlock, opts)

	_, err := twitch.ExecuteRequest[struct{}]("DELETE", url, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] UnblockUser: targetUserID=%v", targetUserID)
		return err
	}

	return nil
}

func GetUserData(login string) (*User, error) {
	opts := map[string]any{
		"login": login,
	}
	url := twitch.BuildURL(urlAPIUsers, opts)

	result, err := twitch.ExecuteRequest[GetUsersResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetUserData: login=%v", login)
		return nil, err
	}

	if len(result.Data) == 0 {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetUserData: login=%v", login)
		return nil, nil
	}

	return &result.Data[0], nil
}

func GetUserDataById(id string) (*User, error) {
	opts := map[string]any{
		"id": id,
	}
	url := twitch.BuildURL(urlAPIUsers, opts)

	result, err := twitch.ExecuteRequest[GetUsersResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetUserDataById: id=%v", id)
		return nil, err
	}

	if len(result.Data) == 0 {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetUserDataById: id=%v", id)
		return nil, nil
	}

	return &result.Data[0], nil
}

func GetUserExtensions() ([]UserExtension, error) {
	url := twitch.BuildURL(urlAPIUserExtensions+"/list", map[string]any{})

	result, err := twitch.ExecuteRequest[GetUserExtensionsResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetUserExtensions: no params")
		return nil, err
	}

	return result.Data, nil
}

func GetUserActiveExtensions() (*UserExtensionData, error) {

	url := twitch.BuildURL(urlAPIUserExtensions, map[string]any{})

	result, err := twitch.ExecuteRequest[UserActiveExtensionsResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetUserActiveExtensions: no params")
		return nil, err
	}

	return &result.Data, nil
}

func UpdateUserExtensions(req UserActiveExtensionsResponse) (*UserActiveExtensionsResponse, error) {
	opts := map[string]any{}
	url := twitch.BuildURL(urlAPIUserExtensions, opts)

	result, err := twitch.ExecuteJSONRequest[UserActiveExtensionsResponse]("PUT", url, req, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] UpdateUserExtensions: request sent")
		return nil, err
	}

	return result, nil
}
