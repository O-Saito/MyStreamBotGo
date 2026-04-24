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
	IsVacationEnabled bool      `json:"is_vacation_enabled"`
	VacationStartTime time.Time `json:"vacation_start_time"`
	VacationEndTime   time.Time `json:"vacation_end_time"`
	Timezone          string    `json:"timezone"`
}

type Schedule struct {
	BroadcasterID       string            `json:"broadcaster_id"`
	BroadcasterName     string            `json:"broadcaster_name"`
	BroadcasterLogin    string            `json:"broadcaster_login"`
	ScheduledBroadcasts []ScheduleSegment `json:"scheduled_broadcasts"`
	ScheduleSettings    ScheduleSettings  `json:"schedule_settings"`
}

type ScheduleData struct {
	Segments         []ScheduleSegment `json:"segments"`
	BroadcasterID    string            `json:"broadcaster_id"`
	BroadcasterName  string            `json:"broadcaster_name"`
	BroadcasterLogin string            `json:"broadcaster_login"`
	Vacation         *VacationInfo     `json:"vacation"`
}

type GetScheduleResponse struct {
	Data       ScheduleData      `json:"data"`
	Pagination twitch.Pagination `json:"pagination"`
}

func GetChannelStreamSchedule(broadcasterID string, ids []string, timezone string, startTime, endTime *time.Time, req *twitch.PaginationRequest) (*twitch.PaginationData[ScheduleData], error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	opts := map[string]any{
		"broadcaster_id": broadcasterID,
	}
	if ids != nil {
		opts["id"] = ids
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

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/schedule", opts)

	result, err := twitch.ExecuteRequest[twitch.PaginationData[ScheduleData]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetChannelStreamSchedule: broadcasterID=%v", broadcasterID)
		return nil, err
	}

	quantity := 0
	if req != nil {
		quantity = req.Quantity
	}
	result.GetNext = func() *twitch.PaginationData[ScheduleData] {
		r, _ := GetChannelStreamSchedule(broadcasterID, ids, timezone, startTime, endTime, &twitch.PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: quantity,
		})
		return r
	}

	return result, nil
}

type UpdateScheduleSegmentRequest struct {
	StartTime  *string `json:"start_time,omitempty"`
	Duration   *string `json:"duration,omitempty"`
	CategoryID *string `json:"category_id,omitempty"`
	Title      *string `json:"title,omitempty"`
	IsCanceled *bool   `json:"is_canceled,omitempty"`
	Timezone   *string `json:"timezone,omitempty"`
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
	Timezone          *string    `json:"timezone,omitempty"`
	IsVacation        *bool      `json:"is_vacation_enabled,omitempty"`
	VacationStartTime *time.Time `json:"vacation_start_time,omitempty"`
	VacationEndTime   *time.Time `json:"vacation_end_time,omitempty"`
}

func UpdateChannelStreamSchedule(req UpdateScheduleSettingsRequest) error {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
	}

	if req.Timezone != nil {
		opts["timezone"] = *req.Timezone
	}
	if req.IsVacation != nil {
		opts["is_vacation_enabled"] = *req.IsVacation
	}
	if req.VacationStartTime != nil {
		opts["vacation_start_time"] = req.VacationStartTime.Format(time.RFC3339)
	}
	if req.VacationEndTime != nil {
		opts["vacation_end_time"] = req.VacationEndTime.Format(time.RFC3339)
	}

	url := twitch.BuildURL(urlAPISchedule, opts)

	_, err := twitch.ExecuteRequest[struct{}]("PATCH", url, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] UpdateChannelStreamSchedule: req=%v", req)
		return err
	}

	return nil
}

type CreateScheduleSegmentRequest struct {
	StartTime   *string `json:"start_time,omitempty"`
	Timezone    *string `json:"timezone,omitempty"`
	IsRecurring *bool   `json:"is_recurring,omitempty"`
	Duration    *string `json:"duration,omitempty"`
	Title       *string `json:"title,omitempty"`
	CategoryID  *string `json:"category_id,omitempty"`
}

type CreateScheduleSegmentResponse struct {
	Data []ScheduleSegment `json:"data"`
}

type VacationInfo struct {
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
}

type UpdateScheduleSegmentResponse struct {
	Data struct {
		Segments         []ScheduleSegment `json:"segments"`
		BroadcasterID    string            `json:"broadcaster_id"`
		BroadcasterName  string            `json:"broadcaster_name"`
		BroadcasterLogin string            `json:"broadcaster_login"`
		Vacation         *VacationInfo     `json:"vacation"`
	} `json:"data"`
}

func CreateChannelStreamScheduleSegment(req CreateScheduleSegmentRequest) (*ScheduleSegment, error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
	}
	url := twitch.BuildURL(twitch.HelixBaseURL+"/schedule/segments", opts)

	result, err := twitch.ExecuteJSONRequest[CreateScheduleSegmentResponse]("POST", url, req, 201)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] CreateChannelStreamScheduleSegment: req:%v", req)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func UpdateChannelStreamScheduleSegment(segmentID string, req UpdateScheduleSegmentRequest) (*ScheduleSegment, error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
		"id":             segmentID,
	}
	url := twitch.BuildURL(twitch.HelixBaseURL+"/schedule/segments", opts)

	result, err := twitch.ExecuteJSONRequest[UpdateScheduleSegmentResponse]("PATCH", url, req, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] UpdateChannelStreamScheduleSegment: segmentID=%v", segmentID)
		return nil, err
	}

	if len(result.Data.Segments) == 0 {
		return nil, nil
	}

	return &result.Data.Segments[0], nil
}

func DeleteChannelStreamScheduleSegment(segmentID string) error {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
		"id":             segmentID,
	}
	url := twitch.BuildURL(twitch.HelixBaseURL+"/schedule/segments", opts)

	_, err := twitch.ExecuteRequest[struct{}]("DELETE", url, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] DeleteChannelStreamScheduleSegment: broadcasterID=%v segmentID=%v", user.UserID, segmentID)
		return err
	}

	return nil
}
