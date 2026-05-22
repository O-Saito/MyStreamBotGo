package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	
	"time"
)

type RaidResponse struct {
	CreatedAt time.Time `json:"created_at"`
	IsMature  bool      `json:"is_mature"`
}

type StartRaidResponse struct {
	Data []RaidResponse `json:"data"`
}

func StartRaid(toBroadcasterID string) (*RaidResponse, error) {
	user := globals.GetState().GetTwitchUser()
	opts := map[string]any{
		"from_broadcaster_id": user.UserID,
		"to_broadcaster_id":   toBroadcasterID,
	}
	url := BuildURL(HelixBaseURL+"/raids", opts)

	result, err := ExecuteRequest[StartRaidResponse]("POST", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] StartRaid: toBroadcasterID=%v", toBroadcasterID)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func CancelRaid() error {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
	}
	url := BuildURL(HelixBaseURL+"/raids", opts)

	_, err := ExecuteRequest[struct{}]("DELETE", url, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] CancelRaid:")
		return err
	}

	return nil
}
