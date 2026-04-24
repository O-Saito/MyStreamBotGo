package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
	"time"
)

var urlAPISchedule = twitch.HelixBaseURL + "/schedule"

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

	opts := map[string]any{
		"broadcaster_id": broadcasterID,
	}
	if timezone != "" {
		opts["timezone"] = timezone
	}
	if startTime != nil {
		opts["start_time"] = startTime.Format(time.RFC3339)
	}
	if endTime != nil {
		opts["end_time"] = endTime.Format(time.RFC3339)
	}

	url := twitch.BuildURL(urlAPISchedule, opts)

	result, err := twitch.ExecuteRequest[GetScheduleResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetChannelStreamSchedule: broadcasterID=%v", broadcasterID)
		return nil, err
	}

	return &result.Data, nil
}

func GetChanneliCalendar(broadcasterID string) (string, error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	opts := map[string]any{
		"broadcaster_id": broadcasterID,
	}

	url := twitch.BuildURL(urlAPISchedule+"/icalendar", opts)

	result, err := twitch.ExecuteRequest[struct {
		Data string `json:"data"`
	}]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetChanneliCalendar: broadcasterID=%v", broadcasterID)
		return "", err
	}

	return result.Data, nil
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

	opts := map[string]any{
		"broadcaster_id": broadcasterID,
	}
	url := twitch.BuildURL(urlAPISchedule, opts)

	_, err := twitch.ExecuteJSONRequest[struct{}, UpdateScheduleSettingsRequest]("PATCH", url, req, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] UpdateChannelStreamSchedule: broadcasterID=%v", broadcasterID)
		return err
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

type CreateScheduleSegmentResponse struct {
	Data []ScheduleSegment `json:"data"`
}

func CreateChannelStreamScheduleSegment(broadcasterID string, req CreateScheduleSegmentRequest) (*ScheduleSegment, error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	opts := map[string]any{
		"broadcaster_id": broadcasterID,
	}
	url := twitch.BuildURL(urlAPISchedule+"/segments", opts)

	result, err := twitch.ExecuteJSONRequest[CreateScheduleSegmentResponse, CreateScheduleSegmentRequest]("POST", url, req, 201)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] CreateChannelStreamScheduleSegment: broadcasterID=%v", broadcasterID)
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

	opts := map[string]any{
		"broadcaster_id": broadcasterID,
		"id":           segmentID,
	}
	url := twitch.BuildURL(urlAPISchedule+"/segments", opts)

	result, err := twitch.ExecuteJSONRequest[CreateScheduleSegmentResponse, UpdateScheduleSegmentRequest]("PATCH", url, req, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] UpdateChannelStreamScheduleSegment: broadcasterID=%v segmentID=%v", broadcasterID, segmentID)
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

	opts := map[string]any{
		"broadcaster_id": broadcasterID,
		"id":             segmentID,
	}
	url := twitch.BuildURL(urlAPISchedule+"/segments", opts)

	_, err := twitch.ExecuteRequest[struct{}]("DELETE", url, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] DeleteChannelStreamScheduleSegment: broadcasterID=%v segmentID=%v", broadcasterID, segmentID)
		return err
	}

	return nil
}