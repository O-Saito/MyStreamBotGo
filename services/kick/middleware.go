package kick

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"net/http"
	"time"
)

// DoRequest executes an HTTP request for Kick with automatic token renewal
// Detects 401 errors (expired token) and attempts to refresh and retry automatically
func DoRequest(req *http.Request) (*http.Response, error) {
	// Set Authorization header
	req.Header.Set("Authorization", "Bearer "+globals.GetState().GetKickUser().Token)

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	// If received 401 (Unauthorized), try automatic refresh
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		helpers.Logf(helpers.DEBUG, "[KICK] Token expired (401), trying refresh...")

		var retry = 0
		err := RefreshAccessToken()
		for err != nil && retry < 3 {
			err = RefreshAccessToken()
			retry++
		}
		if err != nil {
			globals.GetState().SetKickUser(globals.KickUser{})
			return nil, err
		}

		kickUser := globals.GetState().GetKickUser()
		at, rt := kickUser.Token, kickUser.RefreshToken
		exp := kickUser.TokenExpiresIn

		sqlErr := globals.GetGlobalDB().SaveToken("kick", at, rt, time.Now().Add(time.Duration(exp)*time.Second))
		if sqlErr != nil {
			helpers.Logf(helpers.ERROR, "[KICK] Failed to save refreshed token: %s", sqlErr.Error())
		}

		req.Header.Set("Authorization", "Bearer "+at)

		helpers.Logf(helpers.DEBUG, "[KICK] Retrying request with new token")
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}
