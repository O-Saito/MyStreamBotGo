package youtube

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"net/http"
	"time"
)

// DoYouTubeRequest executes an HTTP request for YouTube with automatic token renewal
// Detects 401 errors (expired token) and attempts to refresh and retry automatically
func DoYouTubeRequest(req *http.Request) (*http.Response, error) {
	user := globals.GetState().GetYouTubeUser()

	// Set Authorization header
	req.Header.Set("Authorization", "Bearer "+user.Token)

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	// If received 401 (Unauthorized), try automatic refresh
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		helpers.Logf(helpers.DEBUG, "[YT] Token expired (401), trying refresh...")

		var retry = 0
		err := RefreshToken()
		for err != nil && retry < 3 {
			helpers.Logf(helpers.ERROR, "[YT] Failed to refresh token: %s", err.Error())
			err = RefreshToken()
			retry++
		}
		// Refresh the token
		if err != nil {
			globals.GetState().SetYouTubeUser(globals.YouTubeUser{})
			return nil, err
		}

		// Save new token to database
		newUser := globals.GetState().GetYouTubeUser()
		sqlErr := globals.GetGlobalDB().SaveToken("youtube", newUser.Token, newUser.RefreshToken, time.Now().Add(time.Duration(newUser.TokenExpiresIn)*time.Second))
		if sqlErr != nil {
			helpers.Logf(helpers.ERROR, "[YT] Failed to save refreshed token: %s", sqlErr.Error())
		}

		// Update Authorization header with new token
		req.Header.Set("Authorization", "Bearer "+newUser.Token)

		// Retry again
		helpers.Logf(helpers.DEBUG, "[YT] Retrying request with new token")
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}
