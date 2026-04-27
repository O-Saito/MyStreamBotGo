package twitch

var urlAPIEventSub = HelixBaseURL + "/eventsub/subscriptions"

type EventSubSubscription struct {
	ID        string            `json:"id"`
	Status    string            `json:"status"`
	Type      string            `json:"type"`
	Version   string            `json:"version"`
	Transport EventSubTransport `json:"transport"`
	CreatedAt string            `json:"created_at"`
	Condition EventSubCondition `json:"condition"`
	Cost      int32             `json:"cost"`
}

type EventSubCondition struct {
	UserId              string `json:"user_id"`
	BroadcasterUserId   string `json:"broadcaster_user_id"`
	ModeratorUserId     string `json:"moderator_user_id"`
	ToBroadcasterUserId string `json:"to_broadcaster_user_id"`
}

type EventSubTransport struct {
	Method    string `json:"method"`
	Callback  string `json:"callback"`
	SessionID string `json:"session_id,omitempty"`
	Secret    string `json:"secret,omitempty"`
}

type EventSubData struct {
	PaginationData[EventSubSubscription]
	TotalCost    int `json:"total_cost"`
	MaxTotalCost int `json:"max_total_cost"`
}

func GetEventSubscriptions() (*EventSubData, error) {
	resp, err := ExecuteRequest[EventSubData]("GET", urlAPIEventSub, 200)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func DeleteEventSubscriptions(id string) error {
	url := urlAPIEventSub + "?id=" + id
	_, err := ExecuteRequestNoParse("DELETE", url, 204)
	if err != nil {
		return err
	}
	return nil
}
