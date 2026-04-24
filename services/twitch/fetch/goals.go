package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	
)

type Goal struct {
	ID             string `json:"id"`
	BroadcasterID  string `json:"broadcaster_id"`
	BroadcasterName string `json:"broadcaster_name"`
	BroadcasterLogin string `json:"broadcaster_login"`
	Type           string `json:"type"`
	Description    string `json:"description"`
	CurrentAmount  int    `json:"current_amount"`
	TargetAmount   int    `json:"target_amount"`
	StartedAt      string `json:"started_at"`
	EndsAt         string `json:"ends_at"`
}

type GetCreatorGoalsResponse struct {
	Data []Goal `json:"data"`
}

func GetCreatorGoals() ([]Goal, error) {
	user := globals.GetState().GetTwitchUser()

	url := HelixBaseURL + "/goals?broadcaster_id=" + user.UserID
	result, err := ExecuteRequest[GetCreatorGoalsResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetCreatorGoals: broadcasterID=%v", user.UserID)
		return nil, err
	}

	return result.Data, nil
}