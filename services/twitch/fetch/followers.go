package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	twitch "MyStreamBot/services/twitch"
	"fmt"
	"net/http"
)

type Follower struct {
	UserId     string `json:"user_id"`
	UserName   string `json:"user_name"`
	UserLogin  string `json:"user_login"`
	FollowedAt string `json:"followed_at"`
}

type GetFollowersResponse struct {
	Data       []Follower        `json:"data"`
	Pagination twitch.Pagination `json:"pagination"`
	Total      int               `json:"total"`
}

func GetFollowersData(broadcaster_id, userId string) ([]Follower, error) {
	url := twitch.HelixBaseURL + "/channels/followers"
	if broadcaster_id == "" {
		broadcaster_id = globals.GetState().GetTwitchUser().UserID
	}
	url = fmt.Sprintf("%s?broadcaster_id=%s", url, broadcaster_id)
	if userId != "" {
		url = fmt.Sprintf("%s&user_id=%s", url, userId)
	}

	resp, err := twitch.ExecuteRequest[GetFollowersResponse]("GET", url, http.StatusOK)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] GetFollowersData: broadcaster_id=%v, userId=%v", broadcaster_id, userId)
		return []Follower{}, err
	}
	return resp.Data, nil
}