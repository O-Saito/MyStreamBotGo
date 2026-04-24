package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	
)

type HypeTrain struct {
	ID               string `json:"id"`
	BroadcasterID    string `json:"broadcaster_id"`
	BroadcasterName  string `json:"broadcaster_name"`
	BroadcasterLogin string `json:"broadcaster_login"`
	CooldownEndTime  string `json:"cooldown_end_time"`
	ExpirationTime   string `json:"expiration_time"`
	Goal            HypeTrainGoal `json:"goal"`
	Participants     []HypeTrainParticipant `json:"participants"`
	TopContributions []HypeTrainContribution `json:"top_contributions"`
	TotalContribution int `json:"total_contribution"`
	LastContribution HypeTrainContribution `json:"last_contribution"`
}

type HypeTrainGoal struct {
	ID            string `json:"id"`
	TargetContributions int `json:"target_contributions"`
	TargetPercent int `json:"target_percent"`
	ContributionsSoFar int `json:"contributions_so_far"`
	PercentContributed int `json:"percent_contributed"`
}

type HypeTrainParticipant struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
}

type HypeTrainContribution struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
	Type      string `json:"type"`
	Total     int    `json:"total"`
}

type GetHypeTrainStatusResponse struct {
	Data []HypeTrain `json:"data"`
}

func GetHypeTrainStatus(broadcasterID string) (*HypeTrain, error) {
	user := globals.GetState().GetTwitchUser()
	if broadcasterID == "" {
		broadcasterID = user.UserID
	}

	url := HelixBaseURL + "/hype_train?broadcaster_id=" + broadcasterID

	result, err := ExecuteRequest[GetHypeTrainStatusResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetHypeTrainStatus: broadcasterID=%v", broadcasterID)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}