package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var urlAPISchedule = "https://api.twitch.tv/helix/schedule"

type ScheduleSegment struct {
	ID            string    `json:"id"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Title         string    `json:"title"`
	CanceledUntil time.Time `json:"canceled_until"`
	IsRecurring   bool      `json:"is_recurring"`
	Category      struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"category"`
}

type ScheduleSettings struct {
	IsVacationEnabled   bool      `json:"is_vacation_enabled"`
	VacationStartTime   time.Time `json:"vacation_start_time"`
	VacationEndTime     time.Time `json:"vacation_end_time"`
	Timezone            string    `json:"timezone"`
}

type Schedule struct {
	BroadcasterID           string          `json:"broadcaster_id"`
	BroadcasterName         string          `json:"broadcaster_name"`
	BroadcasterLogin        string          `json:"broadcaster_login"`
	ScheduledBroadcasts    []ScheduleSegment `json:"scheduled_broadcasts"`
	ScheduleSettings       ScheduleSettings `json:"schedule_settings"`
}

type GetScheduleResponse struct {
	Data Schedule `json:"data"`
}

func GetChannelStreamSchedule(broadcasterID, timezone string, startTime, endTime *time.Time) (*Schedule, error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s?broadcaster_id=%s", urlAPISchedule, broadcasterID)
	if timezone != "" {
		url += "&timezone=" + timezone
	}
	if startTime != nil {
		url += "&start_time=" + startTime.Format(time.RFC3339)
	}
	if endTime != nil {
		url += "&end_time=" + endTime.Format(time.RFC3339)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetChannelStreamSchedule http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetChannelStreamSchedule: broadcasterID=%v", broadcasterID)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetChannelStreamSchedule io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("GetChannelStreamSchedule: failed: %s", body)
	}

	var result GetScheduleResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetChannelStreamSchedule io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetChannelStreamSchedule json.Unmarshal failed: %v", err)
		return nil, err
	}

	return &result.Data, nil
}

func GetChanneliCalendar(broadcasterID string) (string, error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s/icalendar?broadcaster_id=%s", urlAPISchedule, broadcasterID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetChanneliCalendar http.NewRequest failed: %v", err)
		return "", err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetChanneliCalendar: broadcasterID=%v", broadcasterID)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetChanneliCalendar io.ReadAll failed: %v", err)
			return "", err
		}
		return "", fmt.Errorf("GetChanneliCalendar: failed: %s", body)
	}

	ical, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetChanneliCalendar io.ReadAll failed: %v", err)
		return "", err
	}

	return string(ical), nil
}

type UpdateScheduleSettingsRequest struct {
	Timezone    *string `json:"timezone,omitempty"`
	IsVacation *bool   `json:"is_vacation_enabled,omitempty"`
	VacationStartTime *time.Time `json:"vacation_start_time,omitempty"`
	VacationEndTime *time.Time `json:"vacation_end_time,omitempty"`
}

func UpdateChannelStreamSchedule(broadcasterID string, req UpdateScheduleSettingsRequest) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	data, err := json.Marshal(req)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateChannelStreamSchedule json.Marshal failed: %v", err)
		return err
	}

	url := fmt.Sprintf("%s?broadcaster_id=%s", urlAPISchedule, broadcasterID)
	httpReq, err := http.NewRequest("PATCH", url, bytes.NewBuffer(data))
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateChannelStreamSchedule http.NewRequest failed: %v", err)
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := twitch.DoRequest(httpReq)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] UpdateChannelStreamSchedule: broadcasterID=%v", broadcasterID)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] UpdateChannelStreamSchedule io.ReadAll failed: %v", err)
			return err
		}
		return fmt.Errorf("UpdateChannelStreamSchedule: failed: %s", body)
	}

	return nil
}

type CreateScheduleSegmentRequest struct {
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Title       string    `json:"title"`
	IsRecurring *bool     `json:"is_recurring,omitempty"`
	CategoryID  *string   `json:"category_id,omitempty"`
}

func CreateChannelStreamScheduleSegment(broadcasterID string, req CreateScheduleSegmentRequest) (*ScheduleSegment, error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	data, err := json.Marshal(req)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] CreateChannelStreamScheduleSegment json.Marshal failed: %v", err)
		return nil, err
	}

	url := fmt.Sprintf("%s/segments?broadcaster_id=%s", urlAPISchedule, broadcasterID)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] CreateChannelStreamScheduleSegment http.NewRequest failed: %v", err)
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := twitch.DoRequest(httpReq)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] CreateChannelStreamScheduleSegment: broadcasterID=%v", broadcasterID)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] CreateChannelStreamScheduleSegment io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("CreateChannelStreamScheduleSegment: failed: %s", body)
	}

	var result struct {
		Data []ScheduleSegment `json:"data"`
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] CreateChannelStreamScheduleSegment io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] CreateChannelStreamScheduleSegment json.Unmarshal failed: %v", err)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

type UpdateScheduleSegmentRequest struct {
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Title     *string    `json:"title,omitempty"`
	CategoryID *string   `json:"category_id,omitempty"`
}

func UpdateChannelStreamScheduleSegment(broadcasterID, segmentID string, req UpdateScheduleSegmentRequest) (*ScheduleSegment, error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	data, err := json.Marshal(req)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateChannelStreamScheduleSegment json.Marshal failed: %v", err)
		return nil, err
	}

	url := fmt.Sprintf("%s/segments?broadcaster_id=%s&id=%s", urlAPISchedule, broadcasterID, segmentID)
	httpReq, err := http.NewRequest("PATCH", url, bytes.NewBuffer(data))
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateChannelStreamScheduleSegment http.NewRequest failed: %v", err)
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := twitch.DoRequest(httpReq)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] UpdateChannelStreamScheduleSegment: broadcasterID=%v, segmentID=%v", broadcasterID, segmentID)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] UpdateChannelStreamScheduleSegment io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("UpdateChannelStreamScheduleSegment: failed: %s", body)
	}

	var result struct {
		Data []ScheduleSegment `json:"data"`
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateChannelStreamScheduleSegment io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] UpdateChannelStreamScheduleSegment json.Unmarshal failed: %v", err)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func DeleteChannelStreamScheduleSegment(broadcasterID, segmentID string) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s/segments?broadcaster_id=%s&id=%s", urlAPISchedule, broadcasterID, segmentID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] DeleteChannelStreamScheduleSegment http.NewRequest failed: %v", err)
		return err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] DeleteChannelStreamScheduleSegment: broadcasterID=%v, segmentID=%v", broadcasterID, segmentID)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] DeleteChannelStreamScheduleSegment io.ReadAll failed: %v", err)
			return err
		}
		return fmt.Errorf("DeleteChannelStreamScheduleSegment: failed: %s", body)
	}

	return nil
}