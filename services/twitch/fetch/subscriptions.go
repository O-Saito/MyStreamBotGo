package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
)

var urlAPISubscriptions = "https://api.twitch.tv/helix/subscriptions"

type Subscription struct {
	BroadcasterID    string `json:"broadcaster_id"`
	BroadcasterLogin string `json:"broadcaster_login"`
	BroadcasterName  string `json:"broadcaster_name"`
	IsGift           bool   `json:"is_gift"`
	GifterLogin      string `json:"gifter_login"`
	GifterName       string `json:"gifter_name"`
	GifterID         string `json:"gifter_id"`
	Tier             string `json:"tier"`
	PlanName         string `json:"plan_name"`
	UserID           string `json:"user_id"`
	UserLogin        string `json:"user_login"`
	UserName         string `json:"user_name"`
	SubscribedAt     string `json:"subscribed_at"`
}

type UserSubscription struct {
	BroadcasterID    string  `json:"broadcaster_id"`
	BroadcasterLogin string  `json:"broadcaster_login"`
	BroadcasterName  string  `json:"broadcaster_name"`
	IsGift           bool    `json:"is_gift"`
	GifterLogin      *string `json:"gifter_login,omitempty"`
	GifterName       *string `json:"gifter_name,omitempty"`
	Tier             string  `json:"tier"`
}

type GetBroadcasterSubscriptionsResponse struct {
	PaginationData[Subscription]
	Points int `json:"points"`
}

type CheckUserSubscriptionResponse struct {
	Data []UserSubscription `json:"data"`
}

func GetBroadcasterSubscriptions(userIDs []string, req *PaginationRequest) (*PaginationData[Subscription], int, error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
	}

	if userIDs != nil {
		opts["user_id"] = userIDs
	}

	if req != nil {
		if req.Cursor != "" {
			opts["after"] = req.Cursor
		}
		if req.Quantity > 0 {
			opts["first"] = req.Quantity
		}
	}

	url := BuildURL(HelixBaseURL+"/subscriptions", opts)

	result, err := ExecuteRequest[GetBroadcasterSubscriptionsResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetBroadcasterSubscriptions: broadcasterID=%v", user.UserID)
		return nil, 0, err
	}

	pagData := &PaginationData[Subscription]{
		Data:       result.Data,
		Pagination: result.Pagination,
		Total:      result.Total,
	}

	pagData.GetNext = func() *PaginationData[Subscription] {
		n, _, _ := GetBroadcasterSubscriptions(userIDs, &PaginationRequest{
			Cursor: result.Pagination.Cursor,
		})
		return n
	}

	return pagData, result.Points, nil
}

func CheckUserSubscription(broadcasterID string) (*UserSubscription, error) {
	user := globals.GetState().GetTwitchUser()
	opts := map[string]any{
		"broadcaster_id": broadcasterID,
		"user_id":        user.UserID,
	}
	url := BuildURL(urlAPISubscriptions+"/user", opts)

	result, err := ExecuteRequest[CheckUserSubscriptionResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] CheckUserSubscription: broadcasterID=%v, userID=%v", broadcasterID, user.UserID)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}
