package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
)

var urlAPICharity = twitch.HelixBaseURL + "/charity/campaigns"

type CharityCampaign struct {
	ID                     string `json:"id"`
	BroadcasterID          string `json:"broadcaster_id"`
	BroadcasterName      string `json:"broadcaster_name"`
	BroadcasterLogin     string `json:"broadcaster_login"`
	CharityName         string `json:"charity_name"`
	Description        string `json:"description"`
	LogoURL             string `json:"logo_url"`
	WebsiteURL         string `json:"website_url"`
	TargetAmount       int     `json:"target_amount"`
	CurrentAmount     int     `json:"current_amount"`
	CurrentAmountCurrency string `json:"current_amount_currency"`
}

type CharityDonation struct {
	ID            string `json:"id"`
	BroadcasterID string `json:"broadcaster_id"`
	UserID       string `json:"user_id"`
	UserName     string `json:"user_name"`
	UserLogin    string `json:"user_login"`
	Amount       int    `json:"amount"`
	Currency     string `json:"currency"`
	Date         string `json:"date"`
}

type GetCharityCampaignResponse struct {
	Data []CharityCampaign `json:"data"`
}

type GetCharityDonationsResponse struct {
	Data       []CharityDonation `json:"data"`
	Pagination twitch.Pagination `json:"pagination"`
}

func GetCharityCampaign() (*CharityCampaign, error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/charity/campaigns", opts)

	result, err := twitch.ExecuteRequest[GetCharityCampaignResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetCharityCampaign: broadcasterID=%v", user.UserID)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func GetCharityCampaignDonations(req *twitch.PaginationRequest) (*twitch.PaginationData[CharityDonation], error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
	}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/charity/donations", opts)

	result, err := twitch.ExecuteRequest[twitch.PaginationData[CharityDonation]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetCharityCampaignDonations: broadcasterID=%v", user.UserID)
		return nil, err
	}

	result.GetNext = func() *twitch.PaginationData[CharityDonation] {
		GetCharityCampaignDonations(&twitch.PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: req.Quantity,
		})
		return result
	}

	return result, nil
}