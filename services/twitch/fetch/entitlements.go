package twitch

import (
	"fmt"
	"net/http"
	"strings"

	twitch "MyStreamBot/services/twitch"
)

var urlAPIEntitlements = twitch.HelixBaseURL + "/entitlements"

type DropEntitlement struct {
	ID              string `json:"id"`
	BeneficiaryID   string `json:"beneficiary_id"`
	CampaignID      string `json:"campaign_id"`
	GameID          string `json:"game_id"`
	EntitlementCode string `json:"entitlement_code"`
	GrantedAt       string `json:"granted_at"`
}

type GetDropEntitlementsResponse struct {
	Data       []DropEntitlement `json:"data"`
	Pagination twitch.Pagination  `json:"pagination"`
}

func GetDropEntitlements(beneficiaryID, campaignID, gameID string, first int, after string) ([]DropEntitlement, error) {
	url := urlAPIEntitlements + "/drops"
	if beneficiaryID != "" {
		url += "?beneficiary_id=" + beneficiaryID
	}
	if campaignID != "" {
		if len(url) > len(urlAPIEntitlements+"/drops") {
			url += "&"
		} else {
			url += "?"
		}
		url += "campaign_id=" + campaignID
	}
	if gameID != "" {
		if !strings.Contains(url, "?") {
			url += "?"
		} else {
			url += "&"
		}
		url += "game_id=" + gameID
	}
	if first > 0 {
		url += fmt.Sprintf("&first=%d", first)
	}
	if after != "" {
		url += "&after=" + after
	}

	resp, err := twitch.ExecuteRequest[GetDropEntitlementsResponse](http.MethodGet, url, http.StatusOK)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

type UpdateDropEntitlementRequest struct {
	EntitlementIDs []string `json:"entitlement_ids"`
	Status         string   `json:"status"`
}

type UpdateDropEntitlementResponse struct{}

func UpdateDropEntitlements(req UpdateDropEntitlementRequest) error {
	url := urlAPIEntitlements + "/drops"
	_, err := twitch.ExecuteJSONRequest[UpdateDropEntitlementResponse, UpdateDropEntitlementRequest](http.MethodPatch, url, req, http.StatusOK)
	if err != nil {
		return err
	}
	return nil
}