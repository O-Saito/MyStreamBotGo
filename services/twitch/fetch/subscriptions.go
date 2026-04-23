package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
)

var urlAPISubscriptions = "https://api.twitch.tv/helix/subscriptions"

type Subscription struct {
	BroadcasterID            string `json:"broadcaster_id"`
	BroadcasterLogin         string `json:"broadcaster_login"`
	BroadcasterName         string `json:"broadcaster_name"`
	IsGift                 bool   `json:"is_gift"`
	GifterLogin            string `json:"gifter_login"`
	GifterName             string `json:"gifter_name"`
	GifterID               string `json:"gifter_id"`
	Tier                   string `json:"tier"`
	PlanName               string `json:"plan_name"`
	UserID                 string `json:"user_id"`
	UserLogin              string `json:"user_login"`
	UserName               string `json:"user_name"`
	SubscribedAt           string `json:"subscribed_at"`
}

type GetBroadcasterSubscriptionsResponse struct {
	Data       []Subscription `json:"data"`
	Pagination Pagination      `json:"pagination"`
}

type CheckUserSubscriptionResponse struct {
	Data []Subscription `json:"data"`
}

func GetBroadcasterSubscriptions(userIDs []string, req *twitch.PaginationRequest) (*twitch.PaginationData[Subscription], error) {
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

	url := twitch.BuildURL(twitch.HelixBaseURL+"/subscriptions", opts)
	for _, id := range userIDs {
		url += "&user_id=" + id
	}

	result, err := twitch.ExecuteRequest[twitch.PaginationData[Subscription]]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetBroadcasterSubscriptions: broadcasterID=%v", user.UserID)
		return nil, err
	}

	result.GetNext = func() *twitch.PaginationData[Subscription] {
		GetBroadcasterSubscriptions(userIDs, &twitch.PaginationRequest{
			Cursor:   result.Pagination.Cursor,
			Quantity: req.Quantity,
		})
		return result
	}

	return result, nil
}

func CheckUserSubscription(broadcasterID, userID string) (*Subscription, error) {
	opts := map[string]any{
		"broadcaster_id": broadcasterID,
		"user_id":       userID,
	}
	url := twitch.BuildURL(urlAPISubscriptions, opts)

	result, err := twitch.ExecuteRequest[CheckUserSubscriptionResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] CheckUserSubscription: broadcasterID=%v, userID=%v", broadcasterID, userID)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}