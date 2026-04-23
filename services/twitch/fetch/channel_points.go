package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"fmt"
	twitch "MyStreamBot/services/twitch"
)

type CustomReward struct {
	ID                                string  `json:"id"`
	Title                             string  `json:"title"`
	Prompt                            string  `json:"prompt"`
	Cost                              int     `json:"cost"`
	IsEnabled                         bool    `json:"is_enabled"`
	BackgroundColor                   string  `json:"background_color"`
	ImageURL                          string  `json:"image_url"`
	DefaultImageURL                   string  `json:"default_image_url"`
	MaxPerStreamSetting               MaxPerStreamSetting `json:"max_per_stream_setting"`
	MaxPerUserPerStreamSetting        MaxPerUserPerStreamSetting `json:"max_per_user_per_stream_setting"`
	GlobalCooldownSetting             GlobalCooldownSetting `json:"global_cooldown_setting"`
	CooldownExpiresAt                 string  `json:"cooldown_expires_at"`
	RedemptionsRedeemedCurrentStream int     `json:"redemptions_redeemed_current_stream"`
	SkipQueue                         bool    `json:"skip_queue"`
}

type MaxPerStreamSetting struct {
	IsEnabled bool `json:"is_enabled"`
	Max       int  `json:"max"`
}

type MaxPerUserPerStreamSetting struct {
	IsEnabled bool `json:"is_enabled"`
	Max       int  `json:"max"`
}

type GlobalCooldownSetting struct {
	IsEnabled           bool `json:"is_enabled"`
	GlobalCooldownSeconds int `json:"global_cooldown_seconds"`
}

type CustomRewardRedemption struct {
	ID              string               `json:"id"`
	BroadcasterID   string               `json:"broadcaster_id"`
	BroadcasterName string               `json:"broadcaster_name"`
	UserID          string               `json:"user_id"`
	UserName        string               `json:"user_name"`
	UserLogin       string               `json:"user_login"`
	UserInput       string               `json:"user_input"`
	Status          string               `json:"status"`
	RedeemedAt      string               `json:"redeemed_at"`
	Reward          CustomRewardRedemptionReward `json:"reward"`
}

type CustomRewardRedemptionReward struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Prompt string `json:"prompt"`
	Cost   int    `json:"cost"`
}

type CreateCustomRewardRequest struct {
	Title                string `json:"title"`
	Prompt               string `json:"prompt,omitempty"`
	Cost                 int    `json:"cost"`
	IsEnabled            *bool  `json:"is_enabled,omitempty"`
	BackgroundColor      string `json:"background_color,omitempty"`
	ImageURL             string `json:"image_url,omitempty"`
	MaxPerStream         *int   `json:"max_per_stream,omitempty"`
	MaxPerUserPerStream  *int   `json:"max_per_user_per_stream,omitempty"`
	GlobalCooldownSeconds *int  `json:"global_cooldown_seconds,omitempty"`
	SkipQueue            *bool  `json:"skip_queue,omitempty"`
}

type UpdateCustomRewardRequest struct {
	Title                string  `json:"title,omitempty"`
	Prompt               string  `json:"prompt,omitempty"`
	Cost                 *int    `json:"cost,omitempty"`
	IsEnabled            *bool   `json:"is_enabled,omitempty"`
	BackgroundColor      string  `json:"background_color,omitempty"`
	ImageURL             string  `json:"image_url,omitempty"`
	MaxPerStream         *int    `json:"max_per_stream,omitempty"`
	MaxPerUserPerStream  *int    `json:"max_per_user_per_stream,omitempty"`
	GlobalCooldownSeconds *int   `json:"global_cooldown_seconds,omitempty"`
	SkipQueue            *bool   `json:"skip_queue,omitempty"`
}

type GetCustomRewardsResponse struct {
	Data []CustomReward `json:"data"`
}

type GetCustomRewardRedemptionsResponse struct {
	Data       []CustomRewardRedemption `json:"data"`
	Pagination twitch.Pagination `json:"pagination"`
}

func CreateCustomReward(req CreateCustomRewardRequest) (*CustomReward, error) {
	user := globals.GetState().GetTwitchUser()

	url := fmt.Sprintf("%s?broadcaster_id=%s", twitch.HelixBaseURL+"/channel_points", user.UserID)
	result, err := twitch.ExecuteJSONRequest[GetCustomRewardsResponse, CreateCustomRewardRequest]("POST", url, req, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] CreateCustomReward: title=%v", req.Title)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("CreateCustomReward: no data returned")
	}

	return &result.Data[0], nil
}

func DeleteCustomReward(rewardID string) error {
	user := globals.GetState().GetTwitchUser()

	url := fmt.Sprintf("%s?broadcaster_id=%s&id=%s", twitch.HelixBaseURL+"/channel_points", user.UserID, rewardID)
	_, err := twitch.ExecuteRequest[struct{}]("DELETE", url, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] DeleteCustomReward: rewardID=%v", rewardID)
		return err
	}

	return nil
}

func GetCustomRewards(onlyManageable bool) ([]CustomReward, error) {
	user := globals.GetState().GetTwitchUser()

	url := fmt.Sprintf("%s?broadcaster_id=%s", twitch.HelixBaseURL+"/channel_points", user.UserID)
	if onlyManageable {
		url += "&manageable=true"
	}

	result, err := twitch.ExecuteRequest[GetCustomRewardsResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetCustomRewards: broadcasterID=%v", user.UserID)
		return nil, err
	}

	return result.Data, nil
}

func UpdateCustomReward(rewardID string, req UpdateCustomRewardRequest) error {
	user := globals.GetState().GetTwitchUser()

	url := fmt.Sprintf("%s?broadcaster_id=%s&id=%s", twitch.HelixBaseURL+"/channel_points", user.UserID, rewardID)
	_, err := twitch.ExecuteJSONRequest[struct{}, UpdateCustomRewardRequest]("PATCH", url, req, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] UpdateCustomReward: rewardID=%v", rewardID)
		return err
	}

	return nil
}

func GetCustomRewardRedemptions(rewardID, status string, req *twitch.PaginationRequest) (*twitch.PaginationData[CustomRewardRedemption], error) {
	user := globals.GetState().GetTwitchUser()

	opts := twitch.RequestOptions{
		BroadcasterID: user.UserID,
		First:     req.Quantity,
		After:     req.Cursor,
	}

	if rewardID != "" {
		opts.UserID = rewardID
	}
	if status != "" {
		opts.ModeratorID = status
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/channel_points/redemptions", opts)

	result, err := twitch.ExecuteRequest[twitch.PaginationData[CustomRewardRedemption]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetCustomRewardRedemptions: rewardID=%v, status=%v", rewardID, status)
		return nil, err
	}

	result.GetNext = func() *twitch.PaginationData[CustomRewardRedemption] {
		GetCustomRewardRedemptions(rewardID, status, &twitch.PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: opts.First,
		})
		return result
	}

	return result, nil
}

func UpdateRedemptionStatus(rewardID, redemptionID string, status string) error {
	user := globals.GetState().GetTwitchUser()

	url := fmt.Sprintf("%s/redemptions?broadcaster_id=%s&reward_id=%s&id=%s", twitch.HelixBaseURL+"/channel_points", user.UserID, rewardID, redemptionID)
	body := map[string]any{"status": status}
	_, err := twitch.ExecuteJSONRequest[struct{}, map[string]any]("PATCH", url, body, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] UpdateRedemptionStatus: rewardID=%v, redemptionID=%v, status=%v", rewardID, redemptionID, status)
		return err
	}

	return nil
}