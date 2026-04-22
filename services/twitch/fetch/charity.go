package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
	"fmt"
	"net/http"
)

var urlAPICharity = twitch.HelixBaseURL + "/charity/campaigns"

type CharityCampaign struct {
	ID             string      `json:"id"`
	BroadcasterID  string      `json:"broadcaster_id"`
	BroadcasterName string     `json:"broadcaster_name"`
	BroadcasterLogin string    `json:"broadcaster_login"`
	CharityName    string      `json:"charity_name"`
	Description    string      `json:"description"`
	LogoURL        string      `json:"logo_url"`
	WebsiteURL     string      `json:"website_url"`
	TargetAmount   int         `json:"target_amount"`
	CurrentAmount  int         `json:"current_amount"`
	CurrentAmountCurrency string `json:"current_amount_currency"`
}

type CharityDonation struct {
	ID              string `json:"id"`
	BroadcasterID   string `json:"broadcaster_id"`
	UserID          string `json:"user_id"`
	UserName        string `json:"user_name"`
	UserLogin       string `json:"user_login"`
	Amount          int    `json:"amount"`
	Currency        string `json:"currency"`
	Date            string `json:"date"`
}

type GetCharityCampaignResponse struct {
	Data []CharityCampaign `json:"data"`
}

type GetCharityDonationsResponse struct {
	Data       []CharityDonation `json:"data"`
	Pagination twitch.Pagination  `json:"pagination"`
}

func GetCharityCampaign(broadcasterID string) (*CharityCampaign, error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s?broadcaster_id=%s", urlAPICharity, broadcasterID)
	result, err := twitch.ExecuteRequest[GetCharityCampaignResponse]("GET", url, http.StatusOK)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetCharityCampaign: broadcasterID=%v", broadcasterID)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func GetCharityCampaignDonations(broadcasterID string, first int, after string) ([]CharityDonation, error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s/donations?broadcaster_id=%s", urlAPICharity, broadcasterID)
	if first > 0 {
		url += fmt.Sprintf("&first=%d", first)
	}
	if after != "" {
		url += "&after=" + after
	}

	result, err := twitch.ExecuteRequest[GetCharityDonationsResponse]("GET", url, http.StatusOK)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetCharityCampaignDonations: broadcasterID=%v", broadcasterID)
		return nil, err
	}

	return result.Data, nil
}