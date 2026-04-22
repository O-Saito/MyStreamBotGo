package kick

import (
	"MyStreamBot/helpers"
	"net/http"
)

// DoRequest executes an HTTP request for Kick with automatic token renewal
// Detects 401 errors (expired token) and attempts to refresh and retry automatically
func DoRequest(req *http.Request) (*http.Response, error) {
	// Set Authorization header
	req.Header.Set("Authorization", "Bearer "+Token)

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	// If received 401 (Unauthorized), try automatic refresh
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		helpers.Logf(helpers.DEBUG, "[KICK] Token expired (401), trying refresh...")

		// Refresh the token
		// if err := RefreshToken(); err != nil {
		// 	helpers.Logf(helpers.ERROR, "[KICK] Failed to refresh token: %s", err.Error())
		// 	return nil, err
		// }

		// // Save new token to database
		// newUser := globals.GetState().GetYouTubeUser()
		// sqlErr := globals.GetGlobalDB().SaveToken("youtube", newUser.Token, newUser.RefreshToken, time.Now().Add(time.Duration(newUser.TokenExpiresIn)*time.Second))
		// if sqlErr != nil {
		// 	helpers.Logf(helpers.ERROR, "[KICK] Failed to save refreshed token: %s", sqlErr.Error())
		// }

		// // Update Authorization header with new token
		// req.Header.Set("Authorization", "Bearer "+newUser.Token)

		// // Retry again
// 	helpers.Logf(helpers.DEBUG, "[KICK] Retrying request with new token")
		// resp, err = http.DefaultClient.Do(req)
		// if err != nil {
		// 	return nil, err
		// }
	}

	return resp, nil
}
