package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var urlAPIPredictions = "https://api.twitch.tv/helix/predictions"

type Prediction struct {
	ID                string            `json:"id"`
	BroadcasterID     string            `json:"broadcaster_id"`
	BroadcasterName   string            `json:"broadcaster_name"`
	BroadcasterLogin  string            `json:"broadcaster_login"`
	Title             string            `json:"title"`
	WinningOutcomeID  string            `json:"winning_outcome_id"`
	Outcomes          []PredictionOutcome `json:"outcomes"`
	Status            string            `json:"status"`
	PredictionWindow int                `json:"prediction_window"`
	Duration          int               `json:"duration"`
	StartedAt         time.Time         `json:"started_at"`
	EndedAt           time.Time         `json:"ended_at"`
	LockedAt          time.Time         `json:"locked_at"`
}

type PredictionOutcome struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Color       string `json:"color"`
	Users       int    `json:"users"`
	ChannelPoints int  `json:"channel_points"`
	TopPredictors []PredictionPredictor `json:"top_predictors"`
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
	PredictionWindow int      `json:"prediction_window"`
	Duration         int      `json:"duration"`
}

func GetPredictions(predictionIDs []string) ([]Prediction, error) {
	user := globals.GetState().GetTwitchUser()

	url := fmt.Sprintf("%s?broadcaster_id=%s", urlAPIPredictions, user.UserID)
	for _, id := range predictionIDs {
		url += fmt.Sprintf("&id=%s", id)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetPredictions http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
	helpers.Logf(helpers.DEBUG, "[TWITCH] GetPredictions: broadcasterID=%v", user.UserID)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetPredictions io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("GetPredictions: failed: %s", body)
	}

	var result GetPredictionsResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetPredictions io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetPredictions json.Unmarshal failed: %v", err)
		return nil, err
	}

	return result.Data, nil
}

func CreatePrediction(req CreatePredictionRequest) (*Prediction, error) {
	user := globals.GetState().GetTwitchUser()

	data, err := json.Marshal(req)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] CreatePrediction json.Marshal failed: %v", err)
		return nil, err
	}

	url := fmt.Sprintf("%s?broadcaster_id=%s", urlAPIPredictions, user.UserID)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] CreatePrediction http.NewRequest failed: %v", err)
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := twitch.DoRequest(httpReq)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] CreatePrediction: broadcasterID=%v, title=%v", user.UserID, req.Title)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] CreatePrediction io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("CreatePrediction: failed: %s", body)
	}

	var result GetPredictionsResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] CreatePrediction io.NewRequest failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] CreatePrediction json.Unmarshal failed: %v", err)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func EndPrediction(predictionID, outcomeID, status string) (*Prediction, error) {
	user := globals.GetState().GetTwitchUser()

	url := fmt.Sprintf("%s?broadcaster_id=%s&id=%s", urlAPIPredictions, user.UserID, predictionID)
	if outcomeID != "" {
		url += "&outcome_id=" + outcomeID
	}
	if status != "" {
		url += "&status=" + status
	}

	req, err := http.NewRequest("PATCH", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] EndPrediction http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] EndPrediction: broadcasterID=%v, predictionID=%v", user.UserID, predictionID)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] EndPrediction io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("EndPrediction: failed: %s", body)
	}

	var result GetPredictionsResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] EndPrediction io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] EndPrediction json.Unmarshal failed: %v", err)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}