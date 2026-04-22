package twitch

import (
	twitch "MyStreamBot/services/twitch"
)

var urlAPIEventSub = twitch.HelixBaseURL + "/eventsub/subscriptions"

type EventSubData struct {
	Subscriptions []EventSubSubscription `json:"data"`
	Total         int                    `json:"total"`
}

type EventSubSubscription struct {
	ID        string            `json:"id"`
	Status    string            `json:"status"`
	Type      string            `json:"type"`
	Version   string            `json:"version"`
	Condition map[string]any    `json:"condition"`
	Transport EventSubTransport `json:"transport"`
	CreatedAt string            `json:"created_at"`
}

type EventSubTransport struct {
	Method    string `json:"method"`
	Callback  string `json:"callback"`
	SessionID string `json:"session_id,omitempty"`
	Secret    string `json:"secret,omitempty"`
}

type GetEventSubscriptionsResponse struct {
	Data       []EventSubSubscription `json:"data"`
	Total      int                    `json:"total"`
	Pagination twitch.Pagination      `json:"pagination"`
}

func GetEventSubscriptions() (*EventSubData, error) {
	resp, err := twitch.ExecuteRequest[GetEventSubscriptionsResponse]("GET", urlAPIEventSub, 200)
	if err != nil {
		return nil, err
	}

	return &EventSubData{
		Subscriptions: resp.Data,
		Total:         resp.Total,
	}, nil
}

func DeleteEventSubscriptions(id string) error {
	url := urlAPIEventSub + "?id=" + id
	_, err := twitch.ExecuteRequestNoParse("DELETE", url, 204)
	if err != nil {
		return err
	}
	return nil
}
