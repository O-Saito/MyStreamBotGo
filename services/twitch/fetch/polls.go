package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	
)

type Poll struct {
	ID                         string      `json:"id"`
	BroadcasterID              string      `json:"broadcaster_id"`
	BroadcasterName            string      `json:"broadcaster_name"`
	BroadcasterLogin           string      `json:"broadcaster_login"`
	Title                      string      `json:"title"`
	Choices                    []PollChoice `json:"choices"`
	BitsVotingEnabled          bool        `json:"bits_voting_enabled"`
	BitsPerVote                int         `json:"bits_per_vote"`
	ChannelPointsVotingEnabled bool        `json:"channel_points_voting_enabled"`
	ChannelPointsPerVote       int         `json:"channel_points_per_vote"`
	Status                     string      `json:"status"`
	Duration                   int         `json:"duration"`
	StartedAt                  string      `json:"started_at"`
	EndedAt                    string      `json:"ended_at"`
}

type PollChoice struct {
	ID                string `json:"id"`
	Title             string `json:"title"`
	Votes             int    `json:"votes"`
	ChannelPointsVotes int   `json:"channel_points_votes"`
	BitsVotes         int    `json:"bits_votes"`
}

type GetPollsResponse struct {
	Data []Poll `json:"data"`
}

type CreatePollRequest struct {
	Title                     string   `json:"title"`
	Choices                   []string `json:"choices"`
	Duration                  int      `json:"duration"`
	BitsVotingEnabled        *bool    `json:"bits_voting_enabled,omitempty"`
	BitsPerVote              *int     `json:"bits_per_vote,omitempty"`
	ChannelPointsVotingEnabled *bool  `json:"channel_points_voting_enabled,omitempty"`
	ChannelPointsPerVote     *int     `json:"channel_points_per_vote,omitempty"`
}

func GetPolls(pollIDs []string) ([]Poll, error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
	}
	for _, id := range pollIDs {
		opts["id"] = id
	}

	url := BuildURL(HelixBaseURL+"/polls", opts)

	result, err := ExecuteRequest[GetPollsResponse]("GET", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetPolls: broadcasterID=%v", user.UserID)
		return nil, err
	}

	return result.Data, nil
}

func CreatePoll(req CreatePollRequest) (*Poll, error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
	}
	url := BuildURL(HelixBaseURL+"/polls", opts)

	result, err := ExecuteJSONRequest[GetPollsResponse, CreatePollRequest]("POST", url, req, 201)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] CreatePoll: broadcasterID=%v, title=%v", user.UserID, req.Title)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func EndPoll(pollID string) (*Poll, error) {
	user := globals.GetState().GetTwitchUser()

	opts := map[string]any{
		"broadcaster_id": user.UserID,
		"id":            pollID,
		"status":        "ended",
	}
	url := BuildURL(HelixBaseURL+"/polls", opts)

	result, err := ExecuteRequest[GetPollsResponse]("PATCH", url, 200)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] EndPoll: broadcasterID=%v, pollID=%v", user.UserID, pollID)
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}