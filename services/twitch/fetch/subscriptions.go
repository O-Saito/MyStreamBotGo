package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var urlAPISubscriptions = "https://api.twitch.tv/helix/subscriptions"

type Subscription struct {
	BroadcasterID             string `json:"broadcaster_id"`
	BroadcasterLogin          string `json:"broadcaster_login"`
	BroadcasterName           string `json:"broadcaster_name"`
	IsGift                    bool   `json:"is_gift"`
	GifterLogin               string `json:"gifter_login"`
	GifterName                string `json:"gifter_name"`
	GifterID                  string `json:"gifter_id"`
	Tier                      string `json:"tier"`
	PlanName                  string `json:"plan_name"`
	UserID                    string `json:"user_id"`
	UserLogin                 string `json:"user_login"`
	UserName                  string `json:"user_name"`
	SubscribedAt              string `json:"subscribed_at"`
}

type GetBroadcasterSubscriptionsResponse struct {
	Data       []Subscription `json:"data"`
	Pagination Pagination     `json:"pagination"`
}

type CheckUserSubscriptionResponse struct {
	Data []Subscription `json:"data"`
}

func GetBroadcasterSubscriptions(broadcasterID string, userIDs []string, req *twitch.PaginationRequest) (*twitch.PaginationData[Subscription], error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	opts := twitch.RequestOptions{
		BroadcasterID: broadcasterID,
	}

	if req != nil {
		opts.After = req.Cursor
		opts.First = req.Quantity
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/subscriptions", opts)
	for _, id := range userIDs {
		url += fmt.Sprintf("&user_id=%s", id)
	}

	result, err := twitch.ExecuteRequest[twitch.PaginationData[Subscription]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetBroadcasterSubscriptions: broadcasterID=%v", broadcasterID)
		return nil, err
	}

	result.GetNext = func() *twitch.PaginationData[Subscription] {
		GetBroadcasterSubscriptions(broadcasterID, userIDs, &twitch.PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: opts.First,
		})
		return result
	}

	return result, nil
}

func CheckUserSubscription(broadcasterID, userID string) (*Subscription, error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}
	if userID == "" {
		userID = user.UserID
	}

	url := fmt.Sprintf("%s?broadcaster_id=%s&user_id=%s", urlAPISubscriptions, broadcasterID, userID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] CheckUserSubscription http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] CheckUserSubscription: broadcasterID=%v, userID=%v", broadcasterID, userID)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] CheckUserSubscription io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("CheckUserSubscription: failed: %s", body)
	}

	var result CheckUserSubscriptionResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] CheckUserSubscription io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] CheckUserSubscription json.Unmarshal failed: %v", err)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}