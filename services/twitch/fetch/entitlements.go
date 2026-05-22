package twitch

import (
	
)

type DropEntitlement struct {
	ID              string `json:"id"`
	BeneficiaryID   string `json:"beneficiary_id"`
	CampaignID      string `json:"campaign_id"`
	GameID          string `json:"game_id"`
	EntitlementCode string `json:"entitlement_code"`
	GrantedAt       string `json:"granted_at"`
}

func GetDropEntitlements(beneficiaryID, campaignID, gameID string, req *PaginationRequest) ([]DropEntitlement, error) {
	opts := map[string]any{}

	if beneficiaryID != "" {
		opts["beneficiary_id"] = beneficiaryID
	}
	if campaignID != "" {
		opts["campaign_id"] = campaignID
	}
	if gameID != "" {
		opts["game_id"] = gameID
	}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := BuildURL(HelixBaseURL+"/entitlements/drops", opts)

	result, err := ExecuteRequest[PaginationData[DropEntitlement]]("GET", url, 200)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

type UpdateDropEntitlementRequest struct {
	EntitlementIDs []string `json:"entitlement_ids"`
	Status         string   `json:"status"`
}

type UpdateDropEntitlementResponse struct{}

func UpdateDropEntitlements(req UpdateDropEntitlementRequest) error {
	url := BuildURL(HelixBaseURL+"/entitlements/drops", nil)

	_, err := ExecuteJSONRequest[UpdateDropEntitlementResponse, UpdateDropEntitlementRequest]("PATCH", url, req, 200)
	if err != nil {
		return err
	}
	return nil
}