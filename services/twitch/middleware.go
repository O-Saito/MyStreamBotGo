package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"fmt"
	"net/http"
)

func AddAuthHeaders(req *http.Request) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", globals.GetState().GetTwitchUser().Token))
	req.Header.Set("Client-ID", globals.GetConfig().TwitchClientID)
}

// DoRequest executes an HTTP request for Twitch with automatic token renewal
// Detects 401 errors (expired token) and attempts to refresh and retry automatically
func DoRequest(req *http.Request) (*http.Response, error) {
	// Set Authorization header
	AddAuthHeaders(req)

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	// If received 401 (Unauthorized), try automatic refresh
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		helpers.Logf(helpers.DEBUG, "[TWITCH] Token expired (401), trying refresh...")

		currentAccess, err := globals.GetGlobalDB().GetToken("twitch")
		_, err = RefreshToken(currentAccess.RefreshToken)
		// Refresh the token
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] Failed to refresh token: %s", err.Error())
			return nil, err
		}

		// Update Authorization header with new token
		AddAuthHeaders(req)

		// Retry again
		helpers.Logf(helpers.DEBUG, "[TWITCH] Retrying request with new token")
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}
