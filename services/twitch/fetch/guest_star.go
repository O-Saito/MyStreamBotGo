package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	
	"fmt"
	"net/http"
)

var urlAPIGuestStar = HelixBaseURL + "/guest_star"

type GuestStarSettings struct {
	BroadcasterID              string `json:"broadcaster_id"`
	IsEnabled                  bool   `json:"is_enabled"`
	MaxGuests                  int    `json:"max_guests"`
	MaxDurationMinutes         int    `json:"max_duration_minutes"`
	AutoSlotAssignmentEnabled bool   `json:"auto_slot_assignment_enabled"`
}

type GuestStarSession struct {
	BroadcasterID    string `json:"broadcaster_id"`
	BroadcasterName  string `json:"broadcaster_name"`
	SessionID       string `json:"session_id"`
	IsLive          bool    `json:"is_live"`
	StartedAt       string `json:"started_at"`
	EndsAt          string `json:"ends_at"`
}

type GuestStarInvite struct {
	ID             string `json:"id"`
	SlotID         string `json:"slot_id"`
	BroadcasterID  string `json:"broadcaster_id"`
	GuestUserID    string `json:"guest_user_id"`
	GuestUserLogin string `json:"guest_user_login"`
	GuestUserName  string `json:"guest_user_name"`
	Status         string `json:"status"`
	InvitedAt      string `json:"invited_at"`
}

type GuestStarSlot struct {
	ID             string `json:"id"`
	SlotIndex     int    `json:"slot_index"`
	GuestUserID    string `json:"guest_user_id"`
	GuestUserLogin string `json:"guest_user_login"`
	GuestUserName  string `json:"guest_user_name"`
	InvitedBy      string `json:"invited_by"`
	AssignedAt     string `json:"assigned_at"`
	ServerSlotID   string `json:"server_slot_id"`
}

type GuestStarSlotSettings struct {
	SlotIndex      int  `json:"slot_index"`
	MuteAudio      bool  `json:"mute_audio"`
	RenderVideo    bool  `json:"render_video"`
	Hidden         bool  `json:"hidden"`
}

type GetGuestStarSettingsResponse struct {
	Data []GuestStarSettings `json:"data"`
}

type GetGuestStarSessionResponse struct {
	Data []GuestStarSession `json:"data"`
}

type GetGuestStarInvitesResponse struct {
	Data []GuestStarInvite `json:"data"`
}

type GuestStarSlotSettingsResponse struct {
	Data []GuestStarSlotSettings `json:"data"`
}

func GetChannelGuestStarSettings(broadcasterID string) (*GuestStarSettings, error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s/settings?broadcaster_id=%s", urlAPIGuestStar, broadcasterID)
	result, err := ExecuteRequest[GetGuestStarSettingsResponse]("GET", url, http.StatusOK)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetChannelGuestStarSettings: broadcasterID=%v", broadcasterID)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

type UpdateGuestStarSettingsRequest struct {
	IsEnabled              *bool `json:"is_enabled,omitempty"`
	MaxGuests              *int  `json:"max_guests,omitempty"`
	MaxDurationMinutes     *int  `json:"max_duration_minutes,omitempty"`
	AutoSlotAssignmentEnabled *bool `json:"auto_slot_assignment_enabled,omitempty"`
}

type EmptyResponse struct{}

func UpdateChannelGuestStarSettings(broadcasterID string, req UpdateGuestStarSettingsRequest) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s/settings?broadcaster_id=%s", urlAPIGuestStar, broadcasterID)
	_, err := ExecuteJSONRequest[EmptyResponse, UpdateGuestStarSettingsRequest]("PATCH", url, req, http.StatusOK)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] UpdateChannelGuestStarSettings: broadcasterID=%v", broadcasterID)
		return err
	}

	return nil
}

func GetGuestStarSession(broadcasterID string) (*GuestStarSession, error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s/sessions?broadcaster_id=%s", urlAPIGuestStar, broadcasterID)
	result, err := ExecuteRequest[GetGuestStarSessionResponse]("GET", url, http.StatusOK)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetGuestStarSession: broadcasterID=%v", broadcasterID)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func CreateGuestStarSession(broadcasterID string) (*GuestStarSession, error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s/sessions?broadcaster_id=%s", urlAPIGuestStar, broadcasterID)
	result, err := ExecuteRequest[GetGuestStarSessionResponse]("POST", url, http.StatusOK)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] CreateGuestStarSession: broadcasterID=%v", broadcasterID)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func EndGuestStarSession(broadcasterID string) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s/sessions?broadcaster_id=%s", urlAPIGuestStar, broadcasterID)
	_, err := ExecuteRequest[EmptyResponse]("DELETE", url, http.StatusNoContent)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] EndGuestStarSession: broadcasterID=%v", broadcasterID)
		return err
	}

	return nil
}

func GetGuestStarInvites(broadcasterID string) ([]GuestStarInvite, error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s/invites?broadcaster_id=%s", urlAPIGuestStar, broadcasterID)
	result, err := ExecuteRequest[GetGuestStarInvitesResponse]("GET", url, http.StatusOK)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetGuestStarInvites: broadcasterID=%v", broadcasterID)
		return nil, err
	}

	return result.Data, nil
}

type SendInviteRequest struct {
	BroadcasterID string `json:"broadcaster_id"`
	GuestUserID   string `json:"guest_user_id"`
}

func SendGuestStarInvite(broadcasterID, guestUserID string) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s/invites?broadcaster_id=%s", urlAPIGuestStar, broadcasterID)
	req := SendInviteRequest{
		BroadcasterID: broadcasterID,
		GuestUserID:   guestUserID,
	}
	_, err := ExecuteJSONRequest[EmptyResponse, SendInviteRequest]("POST", url, req, http.StatusCreated)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] SendGuestStarInvite: broadcasterID=%v, guestUserID=%v", broadcasterID, guestUserID)
		return err
	}

	return nil
}

func DeleteGuestStarInvite(broadcasterID, guestUserID string) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s/invites?broadcaster_id=%s&guest_user_id=%s", urlAPIGuestStar, broadcasterID, guestUserID)
	_, err := ExecuteRequest[EmptyResponse]("DELETE", url, http.StatusNoContent)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] DeleteGuestStarInvite: broadcasterID=%v, guestUserID=%v", broadcasterID, guestUserID)
		return err
	}

	return nil
}

func AssignGuestStarSlot(broadcasterID, guestUserID string) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s/slots?broadcaster_id=%s", urlAPIGuestStar, broadcasterID)
	req := SendInviteRequest{
		BroadcasterID: broadcasterID,
		GuestUserID:   guestUserID,
	}
	_, err := ExecuteJSONRequest[EmptyResponse, SendInviteRequest]("POST", url, req, http.StatusCreated)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] AssignGuestStarSlot: broadcasterID=%v, guestUserID=%v", broadcasterID, guestUserID)
		return err
	}

	return nil
}

type UpdateSlotRequest struct {
	BroadcasterID string             `json:"broadcaster_id"`
	SlotID       string             `json:"slot_id"`
	Settings    GuestStarSlotSettings `json:"settings"`
}

func UpdateGuestStarSlot(broadcasterID, slotID string, req GuestStarSlotSettings) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s/slots?broadcaster_id=%s&slot_id=%s", urlAPIGuestStar, broadcasterID, slotID)
	updateReq := UpdateSlotRequest{
		BroadcasterID: broadcasterID,
		SlotID:       slotID,
		Settings:    req,
	}
	_, err := ExecuteJSONRequest[EmptyResponse, UpdateSlotRequest]("PATCH", url, updateReq, http.StatusNoContent)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] UpdateGuestStarSlot: broadcasterID=%v, slotID=%v", broadcasterID, slotID)
		return err
	}

	return nil
}

func DeleteGuestStarSlot(broadcasterID, slotID string) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s/slots?broadcaster_id=%s&slot_id=%s", urlAPIGuestStar, broadcasterID, slotID)
	_, err := ExecuteRequest[EmptyResponse]("DELETE", url, http.StatusNoContent)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] DeleteGuestStarSlot: broadcasterID=%v, slotID=%v", broadcasterID, slotID)
		return err
	}

	return nil
}

type UpdateSlotSettingsRequest struct {
	BroadcasterID string              `json:"broadcaster_id"`
	Settings      GuestStarSlotSettings `json:"settings"`
}

func UpdateGuestStarSlotSettings(broadcasterID string, req GuestStarSlotSettings) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s/slot_settings?broadcaster_id=%s", urlAPIGuestStar, broadcasterID)
	updateReq := UpdateSlotSettingsRequest{
		BroadcasterID: broadcasterID,
		Settings:   req,
	}
	_, err := ExecuteJSONRequest[EmptyResponse, UpdateSlotSettingsRequest]("PATCH", url, updateReq, http.StatusNoContent)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] UpdateGuestStarSlotSettings: broadcasterID=%v", broadcasterID)
		return err
	}

	return nil
}