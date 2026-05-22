package kick

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	//ClientID     = "01K5HCGTPXM5YWMA214JTZGH4X"
	//ClientSecret = "2eb76a9352af6eca4001b20a5a84d86103467b1c9e5852bf95427b6b511695e6"
	RedirectURI = "http://localhost:1699/kick/callback"
	Scopes      = "user:read channel:read channel:write chat:read chat:write channel:read streamkey:read events:subscribe moderation:ban"
)

func RefreshAccessToken() error {
	current := globals.GetState().GetKickUser()

	data := url.Values{}
	data.Set("client_id", globals.GetConfig().KickClientID)
	data.Set("client_secret", globals.GetConfig().KickClientSecret)
	data.Set("refresh_token", current.RefreshToken)
	data.Set("grant_type", "refresh_token")

	resp, err := http.PostForm("https://id.kick.com/oauth/token", data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helpers.Logf(helpers.ERROR, "[KICK] RefreshToken io.ReadAll failed: %v", err)
		return err
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return err
	}

	if tokenResp.AccessToken != "" {
		globals.GetState().SetKickUser(globals.KickUser{
			Token:          tokenResp.AccessToken,
			RefreshToken:   tokenResp.RefreshToken,
			TokenExpiresIn: tokenResp.ExpiresIn,
			UserID:         current.UserID,
			UserLogin:      current.UserLogin,
		})
		return nil
	}

	return fmt.Errorf("RefreshToken: failed to execute refresh token, response: %s", string(body))
}

func HandleLogin() {
	current := globals.GetState().GetKickUser()
	sqlToken, err := globals.GetGlobalDB().GetToken("kick")
	if err == nil && sqlToken.RefreshToken != "" {
		current.RefreshToken = sqlToken.RefreshToken
	}

	if current.RefreshToken != "" {
		helpers.Print(helpers.Reset, "KICK!")

		globals.GetState().SetKickUser(current)
		err := RefreshAccessToken()
		if err == nil {
			var userData, uErr = GetChannel(0, nil)
			if uErr == nil {
				current.UserID = userData.BroadcasterUserId
				current.UserLogin = userData.Slug

				globals.GetState().SetKickUser(current)
				close(LoginDone)
			} else {
				helpers.Logf(helpers.ERROR, "[KICK] Error getting Kick info: %s", uErr.Error())
			}

		} else {
			helpers.Logf(helpers.ERROR, "[KICK] Failed to refresh token: %s", err.Error())
		}
	}

	http.HandleFunc("/kick/login", func(w http.ResponseWriter, r *http.Request) {
		CodeVerifier = helpers.GenerateRandomString(64)
		codeChallenge := helpers.GenerateCodeChallenge(CodeVerifier)
		OAuthState = helpers.GenerateRandomString(32)

		authURL := fmt.Sprintf(
			"https://id.kick.com/oauth/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=%s&code_challenge=%s&code_challenge_method=S256&state=%s",
			globals.GetConfig().KickClientID,
			url.QueryEscape(RedirectURI),
			url.QueryEscape(Scopes),
			codeChallenge,
			OAuthState,
		)
		http.Redirect(w, r, authURL, http.StatusFound)
	})

	http.HandleFunc("/kick/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != OAuthState {
			http.Error(w, "Invalid state", 400)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Code missing", 400)
			return
		}

		data := url.Values{}
		data.Set("grant_type", "authorization_code")
		data.Set("client_id", globals.GetConfig().KickClientID)
		data.Set("client_secret", globals.GetConfig().KickClientSecret)
		data.Set("redirect_uri", RedirectURI)
		data.Set("code_verifier", CodeVerifier)
		data.Set("code", code)

		resp, err := http.PostForm("https://id.kick.com/oauth/token", data)
		if err != nil {
			http.Error(w, "Token error: "+err.Error(), 500)
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.Logf(helpers.ERROR, "[KICK] login io.ReadAll failed: %v", err)
			http.Error(w, "Error reading response: "+err.Error(), 500)
			return
		}

		helpers.Printf(helpers.Reset, "[KICK LOGIN] TOKEN: %s", body)
		var tokenResp struct {
			AccessToken  string `json:"access_token"`
			TokenType    string `json:"token_type"`
			RefreshToken string `json:"refresh_token"`
			ExpiresIn    int    `json:"expires_in"`
			Scope        string `json:"scope"`
		}
		if err := json.Unmarshal(body, &tokenResp); err != nil {
			helpers.Logf(helpers.ERROR, "[KICK] login json.Unmarshal failed: %v", err)
			http.Error(w, "Error processing response: "+err.Error(), 500)
			return
		}

		helpers.Printf(helpers.Reset, "[KICK LOGIN] Login: access_token: %s; token_type: %s; refresh_token: %s; expires: %d; scope: %s", tokenResp.AccessToken, tokenResp.TokenType, tokenResp.RefreshToken, tokenResp.ExpiresIn, tokenResp.Scope)

		current.Token = tokenResp.AccessToken
		current.RefreshToken = tokenResp.RefreshToken
		current.TokenExpiresIn = tokenResp.ExpiresIn

		globals.GetState().SetKickUser(current)
		var userData, uErr = GetChannel(0, nil)
		if uErr != nil {
			helpers.Logf(helpers.ERROR, "[KICK] Error getting Kick info: %s", uErr.Error())
			http.Error(w, "Error getting user info", 500)
			return
		}

		current.UserID = userData.BroadcasterUserId
		current.UserLogin = userData.Slug

		close(LoginDone)

		sqlErr := globals.GetGlobalDB().SaveToken("kick", current.Token, current.RefreshToken, time.Now().Add(time.Duration(current.TokenExpiresIn)*time.Second))

		if sqlErr != nil {
			helpers.Logf(helpers.ERROR, "[KICK HANDLELOGIN] Failed to save token: %s", sqlErr.Error())
		}
		fmt.Fprintf(w, "Login Kick completed! You may close this page.\r\n")
		helpers.Printf(helpers.Reset, "[KICK LOGIN] Login completed: %s (ID: %d)", current.UserLogin, current.UserID)

		if err := Connect(); err != nil {
			helpers.Logf(helpers.ERROR, "[KICK] Error connecting: %s", err.Error())
		}
	})
}
