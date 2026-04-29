package twitch

import (
	"MyStreamBot/globals"
	"fmt"
)

type AdScheduleData struct {
	SnoozeCount     int    `json:"snooze_count"`
	SnoozeRefreshAt string `json:"snooze_refresh_at"`
	NextAdAt        string `json:"next_ad_at"`
	Duration        int    `json:"duration"`
	LastAdAt        string `json:"last_ad_at"`
	PrerollFreeTime int    `json:"preroll_free_time"`
}

type StartCommercialResponse struct {
	Length     int    `json:"length"`
	Message    string `json:"message"`
	RetryAfter int    `json:"retry_after"`
}

type StartCommercialResponseWrapper struct {
	Data []StartCommercialResponse `json:"data"`
}

type SnoozeAdResponse struct {
	SnoozeCount     int    `json:"snooze_count"`
	SnoozeRefreshAt string `json:"snooze_refresh_at"`
	NextAdAt        string `json:"next_ad_at"`
}

type SnoozeAdResponseWrapper struct {
	Data []SnoozeAdResponse `json:"data"`
}

type AdScheduleResponseWrapper struct {
	Data []AdScheduleData `json:"data"`
}

func StartCommercial(broadcasterID string, length int) (*StartCommercialResponse, error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	reqBody := struct {
		BroadcasterID string `json:"broadcaster_id"`
		Length        int    `json:"length"`
	}{
		BroadcasterID: broadcasterID,
		Length:        length,
	}

	url := HelixBaseURL + "/channels/commercial"
	result, err := ExecuteJSONRequest[StartCommercialResponseWrapper]("POST", url, reqBody, 200)
	if err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("StartCommercial: no data returned")
	}

	return &result.Data[0], nil
}

func GetAdSchedule(broadcasterID string) (*AdScheduleData, error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s/channels/ads?broadcaster_id=%s", HelixBaseURL, broadcasterID)
	result, err := ExecuteRequest[AdScheduleResponseWrapper]("GET", url, 200)
	if err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func SnoozeNextAd(broadcasterID string) (*SnoozeAdResponse, error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s/channels/ads/schedule/snooze?broadcaster_id=%s", HelixBaseURL, broadcasterID)
	result, err := ExecuteRequest[SnoozeAdResponseWrapper]("POST", url, 200)
	if err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}
