package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

type RequestOptions struct {
	UserID        string
	BroadcasterID string
	ModeratorID   string
	First         int
	After         string
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

func BuildURL(base string, opts RequestOptions) string {
	url := base
	hasParams := false

	if opts.UserID != "" {
		url += fmt.Sprintf("%suser_id=%s", ternary(hasParams, "&", "?"), opts.UserID)
		hasParams = true
	}
	if opts.BroadcasterID != "" {
		url += fmt.Sprintf("%sbroadcaster_id=%s", ternary(hasParams, "&", "?"), opts.BroadcasterID)
		hasParams = true
	}
	if opts.ModeratorID != "" {
		url += fmt.Sprintf("%smoderator_id=%s", ternary(hasParams, "&", "?"), opts.ModeratorID)
		hasParams = true
	}
	if opts.First > 0 {
		url += fmt.Sprintf("%sfirst=%d", ternary(hasParams, "&", "?"), opts.First)
		hasParams = true
	}
	if opts.After != "" {
		url += fmt.Sprintf("%safter=%s", ternary(hasParams, "&", "?"), opts.After)
		hasParams = true
	}

	return url
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

func getDefaultBroadcasterID() string {
	return globals.GetState().GetTwitchUser().UserID
}

func getRequestOpts(opts RequestOptions) RequestOptions {
	if opts.BroadcasterID == "" {
		opts.BroadcasterID = getDefaultBroadcasterID()
	}
	if opts.ModeratorID == "" {
		opts.ModeratorID = getDefaultBroadcasterID()
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
