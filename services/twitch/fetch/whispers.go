package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
	"fmt"
	"io"
	"net/http"
)

var urlAPIWhispers = "https://api.twitch.tv/helix/whispers"

func SendWhisper(fromUserID, toUserID, message string) error {
	user := globals.GetState().GetTwitchUser()
	if fromUserID == "" {
		fromUserID = user.UserID
	}

	url := fmt.Sprintf("%s?from_user_id=%s&to_user_id=%s", urlAPIWhispers, fromUserID, toUserID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] SendWhisper http.NewRequest failed: %v", err)
		return err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] SendWhisper: fromUserID=%v, toUserID=%v", fromUserID, toUserID)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] SendWhisper io.ReadAll failed: %v", err)
			return err
		}
		return fmt.Errorf("SendWhisper: failed: %s", body)
	}

	return nil
}