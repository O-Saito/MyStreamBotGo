package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
)

var urlAPIRaids = twitch.HelixBaseURL + "/raids"

func StartRaid(fromBroadcasterID, toBroadcasterID string) error {
	user := globals.GetState().GetTwitchUser()
	if fromBroadcasterID == "" {
		fromBroadcasterID = user.UserID
	}

	opts := map[string]any{
		"from_broadcaster_id": fromBroadcasterID,
		"to_broadcaster_id":   toBroadcasterID,
	}
	url := twitch.BuildURL(urlAPIRaids, opts)

	_, err := twitch.ExecuteRequest[struct{}]("POST", url, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] StartRaid: fromBroadcasterID=%v toBroadcasterID=%v", fromBroadcasterID, toBroadcasterID)
		return err
	}

	return nil
}

func CancelRaid(broadcasterID string) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	opts := map[string]any{
		"broadcaster_id": broadcasterID,
	}
	url := twitch.BuildURL(urlAPIRaids, opts)

	_, err := twitch.ExecuteRequest[struct{}]("DELETE", url, 204)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] CancelRaid: broadcasterID=%v", broadcasterID)
		return err
	}

	return nil
}