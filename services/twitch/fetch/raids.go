package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
	"fmt"
	"io"
	"net/http"
)

var urlAPIRaids = "https://api.twitch.tv/helix/raids"

func StartRaid(fromBroadcasterID, toBroadcasterID string) error {
	user := globals.GetState().GetTwitchUser()
	if fromBroadcasterID == "" {
		fromBroadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s?from_broadcaster_id=%s&to_broadcaster_id=%s", urlAPIRaids, fromBroadcasterID, toBroadcasterID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] StartRaid http.NewRequest failed: %v", err)
		return err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] StartRaid: fromBroadcasterID=%v, toBroadcasterID=%v", fromBroadcasterID, toBroadcasterID)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] StartRaid io.ReadAll failed: %v", err)
			return err
		}
		return fmt.Errorf("StartRaid: failed: %s", body)
	}

	return nil
}

func CancelRaid(broadcasterID string) error {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := fmt.Sprintf("%s?broadcaster_id=%s", urlAPIRaids, broadcasterID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] CancelRaid http.NewRequest failed: %v", err)
		return err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] CancelRaid: broadcasterID=%v", broadcasterID)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] CancelRaid io.ReadAll failed: %v", err)
			return err
		}
		return fmt.Errorf("CancelRaid: failed: %s", body)
	}

	return nil
}