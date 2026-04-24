package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	HelixBaseURL = "https://api.twitch.tv/helix"
	IDBaseURL    = "https://id.twitch.tv"
)

type PaginationData[Data any] struct {
	Data       []Data     `json:"data"`
	Pagination Pagination `json:"pagination"`
	Total      int        `json:"total"`
	GetNext    func() *PaginationData[Data]
}

type Pagination struct {
	Cursor string `json:"cursor"`
}

type PaginationRequest struct {
	Cursor   string
	Quantity int
}

type DateRange struct {
	StartedAt string `json:"started_at"`
	EndedAt   string `json:"ended_at"`
}

type TwitchViewerData struct {
	UserId     string `json:"user_id"`
	UserName   string `json:"user_name"`
	UserLogin  string `json:"user_login"`
	FollowedAt string `json:"followed_at"`
}

func AddAuthHeaders(req *http.Request) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", globals.GetState().GetTwitchUser().Token))
	req.Header.Set("Client-ID", globals.GetConfig().TwitchClientID)
}

func DoRequest(req *http.Request) (*http.Response, error) {
	AddAuthHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		helpers.Logf(helpers.DEBUG, "[TWITCH] Token expired (401), trying refresh...")

		currentAccess, err := globals.GetGlobalDB().GetToken("twitch")
		_, err = RefreshToken(currentAccess.RefreshToken)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] Failed to refresh token: %s", err.Error())
			return nil, err
		}

		AddAuthHeaders(req)

		helpers.Logf(helpers.DEBUG, "[TWITCH] Retrying request with new token")
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}

func BuildURL(base string, opts map[string]any) string {
	var url strings.Builder
	url.WriteString(base)
	hasParams := false

	for k, v := range opts {

		if v == nil || v == "" {
			continue
		}

		strArr, ok := v.([]string)
		if ok {
			if len(strArr) == 0 {
				continue
			}
			for _, d := range strArr {
				url.WriteString(fmt.Sprintf("%s%s=%s", ternary(hasParams, "&", "?"), k, d))
				hasParams = true
			}
			continue
		}
		intArr, ok := v.([]int)
		if ok {
			if len(intArr) == 0 {
				continue
			}
			for _, d := range intArr {
				url.WriteString(fmt.Sprintf("%s%s=%s", ternary(hasParams, "&", "?"), k, d))
				hasParams = true
			}
			continue
		}

		url.WriteString(fmt.Sprintf("%s%s=%s", ternary(hasParams, "&", "?"), k, v))
		hasParams = true
	}

	return url.String()
}

func AddIDParam(base, paramName string, id string) string {
	if id == "" {
		return base
	}
	connector := ternary(contains(base, "?"), "&", "?")
	return fmt.Sprintf("%s%s%s=%s", base, connector, paramName, id)
}

func AddIDsParam(base, paramName string, ids []string) string {
	if len(ids) == 0 {
		return base
	}
	connector := ternary(contains(base, "?"), "&", "?")
	result := base
	for _, id := range ids {
		result += fmt.Sprintf("%s%s=%s", connector, paramName, id)
		connector = "&"
	}
	return result
}

func ExecuteRequest[T any](method, url string, expectedStatus int) (*T, error) {
	var req *http.Request
	var err error

	if method == "POST" || method == "PATCH" || method == "PUT" {
		req, err = http.NewRequest(method, url, nil)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] executeRequest http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] executeRequest: url=%v", url)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] executeRequest io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("executeRequest(%s %s): expected %d, got %d: %s", method, url, expectedStatus, resp.StatusCode, respBody)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] executeRequest io.ReadAll failed: %v", err)
		return nil, err
	}

	var result T
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &result); err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] executeRequest json.Unmarshal failed: %v", err)
			return nil, err
		}
	}

	return &result, nil
}

func ExecuteJSONRequest[Resp any, Req any](method, url string, body Req, expectedStatus int) (*Resp, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] executeJSONRequest json.Marshal failed: %v", err)
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] executeJSONRequest http.NewRequest failed: %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] executeJSONRequest: url=%v", url)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] executeJSONRequest io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("executeJSONRequest(%s %s): expected %d, got %d: %s", method, url, expectedStatus, resp.StatusCode, respBody)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] executeJSONRequest io.ReadAll failed: %v", err)
		return nil, err
	}

	var result Resp
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &result); err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] executeJSONRequest json.Unmarshal failed: %v", err)
			return nil, err
		}
	}

	return &result, nil
}

func ExecuteRequestNoParse(method, url string, expectedStatus int) ([]byte, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] executeRequestNoParse http.NewRequest failed: %v", err)
		return nil, err
	}

	resp, err := DoRequest(req)
	if err != nil {
		helpers.Logf(helpers.DEBUG, "[TWITCH] executeRequestNoParse: url=%v", url)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] executeRequestNoParse io.ReadAll failed: %v", err)
			return nil, err
		}
		return nil, fmt.Errorf("executeRequestNoParse(%s %s): expected %d, got %d: %s", method, url, expectedStatus, resp.StatusCode, respBody)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] executeRequestNoParse io.ReadAll failed: %v", err)
		return nil, err
	}

	return respBody, nil
}

func RefreshToken(refreshToken string) (*struct {
	AccessToken  string `json:"access_token"`
	Expires      int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}, error) {
	data := url.Values{}
	data.Set("client_id", globals.GetConfig().TwitchClientID)
	data.Set("client_secret", globals.GetConfig().TwitchClientSecret)
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	resp, err := http.PostForm("https://id.twitch.tv/oauth2/token", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] login io.ReadAll failed: %v", err)
		return nil, err
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		Expires      int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] login json.Unmarshal failed: %v", err)
		return nil, err
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("[TWITCH] empty access token")
	}

	sqlErr := globals.GetGlobalDB().SaveToken("twitch", tokenResp.AccessToken, tokenResp.RefreshToken, time.Now().Add(time.Duration(tokenResp.Expires)*time.Second))
	if sqlErr != nil {
		helpers.Logf(helpers.ERROR, "[TWITCH] Failed to save token: %s", sqlErr.Error())
	}

	user := globals.GetState().GetTwitchUser()
	user.Token = tokenResp.AccessToken
	globals.GetState().SetTwitchUser(user)

	return &tokenResp, nil
}

func getDefaultBroadcasterID() string {
	return globals.GetState().GetTwitchUser().UserID
}

func getRequestOpts(opts *map[string]any) *map[string]any {
	if (*opts)["broadcaster_id"] == "" {
		(*opts)["broadcaster_id"] = getDefaultBroadcasterID()
	}
	if (*opts)["moderator_id"] == "" {
		(*opts)["moderator_id"] = getDefaultBroadcasterID()
	}
	return opts
}

func ternary[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
