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
	ViewCount       int    `json:"view_count"` // decrapted
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

type GetUserExtensionsResponse struct {
	Data []UserExtension `json:"data"`
}

type GetUserAuthorizationResponse struct {
	Data []UserAuthorization `json:"data"`
}

func GetUser(id, login []string) (*User, error) {
	opts := map[string]any{
		"id":    id,
		"login": login,
	}
	url := twitch.BuildURL(urlAPIUsers, opts)

	result, err := twitch.ExecuteRequest[GetUsersResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetUser: id=%v", id)
		return nil, err
	}

	if len(result.Data) == 0 {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetUser: id=%v", id)
		return nil, nil
	}

	return &result.Data[0], nil
}

func UpdateUser(description string) (*User, error) {
	opts := map[string]any{"description": description}
	url := twitch.BuildURL(urlAPIUsers, opts)

	body := map[string]any{}
	result, err := twitch.ExecuteJSONRequest[GetUsersResponse]("PUT", url, body, 200)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateUser: description=%v", description)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func GetAuthorizationByUser(userIDs []string) (*UserAuthorization, error) {
	opts := map[string]any{
		"iser_id": userIDs,
	}
	url := twitch.BuildURL(twitch.HelixBaseURL+"authorization/users", opts)

	result, err := twitch.ExecuteRequest[GetUserAuthorizationResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetAuthorizationByUser: userID=%v", userIDs)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func GetUserBlockList(targetUserID string, req *twitch.PaginationRequest) (*twitch.PaginationData[UserBlock], error) {
	opts := map[string]any{
		"target_user_id": targetUserID,
	}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := twitch.BuildURL(urlAPIUserBlock, opts)

	result, err := twitch.ExecuteRequest[twitch.PaginationData[UserBlock]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetUserBlockList: targetUserID=%v", targetUserID)
		return nil, err
	}

	result.GetNext = func() *twitch.PaginationData[UserBlock] {
		quantity := 0
		if req != nil {
			quantity = req.Quantity
		}
		GetUserBlockList(targetUserID, &twitch.PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: quantity,
		})
		return result
	}

	return result, nil
}

func BlockUser(targetUserID string, sourceContext string, reason string) error {
	opts := map[string]any{
		"target_user_id": targetUserID,
	}
	if sourceContext != "" {
		opts["source_context"] = sourceContext
	}
	if reason != "" {
		opts["reason"] = reason
	}
	url := twitch.BuildURL(urlAPIUserBlock, opts)

	_, err := twitch.ExecuteRequest[struct{}]("PUT", url, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] BlockUser: targetUserID=%v sourceContext=%v reason=%v", targetUserID, sourceContext, reason)
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
