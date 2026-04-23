package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
)

type Prediction struct {
	ID               string               `json:"id"`
	BroadcasterID    string               `json:"broadcaster_id"`
	BroadcasterName  string               `json:"broadcaster_name"`
	BroadcasterLogin string               `json:"broadcaster_login"`
	Title            string               `json:"title"`
	WinningOutcomeID string               `json:"winning_outcome_id"`
	Outcomes         []PredictionOutcome `json:"outcomes"`
	Status           string               `json:"status"`
	PredictionWindow int                  `json:"prediction_window"`
	Duration         int                  `json:"duration"`
	StartedAt        string               `json:"started_at"`
	EndedAt          string               `json:"ended_at"`
	LockedAt         string               `json:"locked_at"`
}

type PredictionOutcome struct {
	ID             string                 `json:"id"`
	Title          string                 `json:"title"`
	Color          string                 `json:"color"`
	Users          int                    `json:"users"`
	ChannelPoints  int                    `json:"channel_points"`
	TopPredictors  []PredictionPredictor `json:"top_predictors"`
}

type PredictionPredictor struct {
	UserID            string `json:"user_id"`
	UserLogin         string `json:"user_login"`
	UserName          string `json:"user_name"`
	ChannelPointsUsed int    `json:"channel_points_used"`
	ChannelPointsWon  int    `json:"channel_points_won"`
}

type GetPredictionsResponse struct {
	Data []Prediction `json:"data"`
}

type CreatePredictionRequest struct {
	Title            string   `json:"title"`
	Outcomes         []string `json:"outcomes"`
	PredictionWindow int     `json:"prediction_window"`
	Duration         int     `json:"duration"`
}

func GetPredictions(predictionIDs []string) ([]Prediction, error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
	}
	for _, id := range predictionIDs {
		opts["id"] = id
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/predictions", opts)

	result, err := twitch.ExecuteRequest[GetPredictionsResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetPredictions: broadcasterID=%v", user.UserID)
		return nil, err
	}

	return result.Data, nil
}

func CreatePrediction(req CreatePredictionRequest) (*Prediction, error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
	}
	url := twitch.BuildURL(twitch.HelixBaseURL+"/predictions", opts)

	result, err := twitch.ExecuteJSONRequest[GetPredictionsResponse, CreatePredictionRequest]("POST", url, req, 201)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] CreatePrediction: broadcasterID=%v, title=%v", user.UserID, req.Title)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func EndPrediction(predictionID, outcomeID, status string) (*Prediction, error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
		"id":            predictionID,
	}
	if outcomeID != "" {
		opts["outcome_id"] = outcomeID
	}
	if status != "" {
		opts["status"] = status
	}

	url := twitch.BuildURL(twitch.HelixBaseURL+"/predictions", opts)

	result, err := twitch.ExecuteRequest[GetPredictionsResponse]("PATCH", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] EndPrediction: broadcasterID=%v, predictionID=%v", user.UserID, predictionID)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}