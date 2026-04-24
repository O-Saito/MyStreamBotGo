package twitch

import (
	"MyStreamBot/helpers"
)

var urlAPIUsers = HelixBaseURL + "/users"
var urlAPIUserBlock = HelixBaseURL + "/users/block"
var urlAPIUserExtensions = HelixBaseURL + "/users/extensions"

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
	url := BuildURL(urlAPIUsers, opts)

	result, err := ExecuteRequest[GetUsersResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetUser: id=%v login=%v url=%v error=%v", id, login, url, err)
		return nil, err
	}

	if len(result.Data) == 0 {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetUser: id=%v login=%v url=%v error=not found", id, login, url)
		return nil, nil
	}

	return &result.Data[0], nil
}

func UpdateUser(description string) (*User, error) {
	opts := map[string]any{"description": description}
	url := BuildURL(urlAPIUsers, opts)

	body := map[string]any{}
	result, err := ExecuteJSONRequest[GetUsersResponse]("PUT", url, body, 200)
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
	url := BuildURL(HelixBaseURL+"authorization/users", opts)

	result, err := ExecuteRequest[GetUserAuthorizationResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetAuthorizationByUser: userID=%v error=%v", userIDs, err)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func GetUserBlockList(targetUserID string, req *PaginationRequest) (*PaginationData[UserBlock], error) {
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

	url := BuildURL(urlAPIUserBlock, opts)

	result, err := ExecuteRequest[PaginationData[UserBlock]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetUserBlockList: targetUserID=%v error=%v", targetUserID, err)
		return nil, err
	}

	result.GetNext = func() *PaginationData[UserBlock] {
		quantity := 0
		if req != nil {
			quantity = req.Quantity
		}
		n, _ := GetUserBlockList(targetUserID, &PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: quantity,
		})
		return n
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
	url := BuildURL(urlAPIUserBlock, opts)

	_, err := ExecuteRequest[struct{}]("PUT", url, 204)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] BlockUser: targetUserID=%v sourceContext=%v reason=%v error=%v", targetUserID, sourceContext, reason, err)
		return err
	}

	return nil
}

func UnblockUser(targetUserID string) error {
	opts := map[string]any{
		"target_user_id": targetUserID,
	}
	url := BuildURL(urlAPIUserBlock, opts)

	_, err := ExecuteRequest[struct{}]("DELETE", url, 204)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UnblockUser: targetUserID=%v error=%v", targetUserID, err)
		return err
	}

	return nil
}

func GetUserExtensions() ([]UserExtension, error) {
	url := BuildURL(urlAPIUserExtensions+"/list", map[string]any{})

	result, err := ExecuteRequest[GetUserExtensionsResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetUserExtensions: error=%v", err)
		return nil, err
	}

	return result.Data, nil
}

func GetUserActiveExtensions() (*UserExtensionData, error) {

	url := BuildURL(urlAPIUserExtensions, map[string]any{})

	result, err := ExecuteRequest[UserActiveExtensionsResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetUserActiveExtensions: error=%v", err)
		return nil, err
	}

	return &result.Data, nil
}

func UpdateUserExtensions(req UserActiveExtensionsResponse) (*UserActiveExtensionsResponse, error) {
	opts := map[string]any{}
	url := BuildURL(urlAPIUserExtensions, opts)

	result, err := ExecuteJSONRequest[UserActiveExtensionsResponse]("PUT", url, req, 200)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateUserExtensions: error=%v", err)
		return nil, err
	}

	return result, nil
}
