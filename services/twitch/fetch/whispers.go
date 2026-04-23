package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
)

func SendWhisper(toUserID, message string) error {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"from_user_id": user.UserID,
		"to_user_id":   toUserID,
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/whispers", opts)

	body := map[string]any{"message": message}
	_, err := twitch.ExecuteJSONRequest[struct{}]("POST", url, body, 204)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] SendWhisper failed: %v", err)
		return err
	}

	return nil
}
