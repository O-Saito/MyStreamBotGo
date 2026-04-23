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

var urlAPIPolls = "https://api.twitch.tv/helix/polls"

type Poll struct {
	ID                string      `json:"id"`
	BroadcasterID     string      `json:"broadcaster_id"`
	BroadcasterName   string      `json:"broadcaster_name"`
	BroadcasterLogin  string      `json:"broadcaster_login"`
	Title             string      `json:"title"`
	Choices           []PollChoice `json:"choices"`
	BitsVotingEnabled bool        `json:"bits_voting_enabled"`
	BitsPerVote       int         `json:"bits_per_vote"`
	ChannelPointsVotingEnabled bool `json:"channel_points_voting_enabled"`
	ChannelPointsPerVote int      `json:"channel_points_per_vote"`
	Status            string      `json:"status"`
	Duration          int         `json:"duration"`
	StartedAt         time.Time   `json:"started_at"`
	EndedAt           time.Time   `json:"ended_at"`
}

type PollChoice struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Votes        int    `json:"votes"`
	ChannelPointsVotes int `json:"channel_points_votes"`
	BitsVotes    int    `json:"bits_votes"`
}

type GetPollsResponse struct {
	Data []Poll `json:"data"`
}

type CreatePollRequest struct {
	Title                    string            `json:"title"`
	Choices                  []string          `json:"choices"`
	Duration                 int               `json:"duration"`
	BitsVotingEnabled        *bool             `json:"bits_voting_enabled,omitempty"`
	BitsPerVote              *int              `json:"bits_per_vote,omitempty"`
	ChannelPointsVotingEnabled *bool           `json:"channel_points_voting_enabled,omitempty"`
	ChannelPointsPerVote     *int              `json:"channel_points_per_vote,omitempty"`
}

func GetPolls(pollIDs []string) ([]Poll, error) {
	user := globals.GetState().GetTwitchUser()

	url := fmt.Sprintf("%s?broadcaster_id=%s", urlAPIPolls, user.UserID)
	for _, id := range pollIDs {
		url += fmt.Sprintf("&id=%s", id)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetPolls http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetPolls: broadcasterID=%v", user.UserID)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] GetPolls io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("GetPolls: failed: %s", body)
	}

	var result GetPollsResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetPolls io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] GetPolls json.Unmarshal failed: %v", err)
		return nil, err
	}

	return result.Data, nil
}

func CreatePoll(req CreatePollRequest) (*Poll, error) {
	user := globals.GetState().GetTwitchUser()

	data, err := json.Marshal(req)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] CreatePoll json.Marshal failed: %v", err)
		return nil, err
	}

	url := fmt.Sprintf("%s?broadcaster_id=%s", urlAPIPolls, user.UserID)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] CreatePoll http.NewRequest failed: %v", err)
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := twitch.DoRequest(httpReq)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] CreatePoll: broadcasterID=%v, title=%v", user.UserID, req.Title)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] CreatePoll io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("CreatePoll: failed: %s", body)
	}

	var result GetPollsResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] CreatePoll io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] CreatePoll json.Unmarshal failed: %v", err)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func EndPoll(pollID string) (*Poll, error) {
	user := globals.GetState().GetTwitchUser()

	url := fmt.Sprintf("%s?broadcaster_id=%s&id=%s&status=ended", urlAPIPolls, user.UserID, pollID)
	req, err := http.NewRequest("PATCH", url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] EndPoll http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := twitch.DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] EndPoll: broadcasterID=%v, pollID=%v", user.UserID, pollID)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] EndPoll io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("EndPoll: failed: %s", body)
	}

	var result GetPollsResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] EndPoll io.ReadAll failed: %v", err)
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] EndPoll json.Unmarshal failed: %v", err)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}