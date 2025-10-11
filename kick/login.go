package kick

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

const (
	//ClientID     = "01K5HCGTPXM5YWMA214JTZGH4X"
	//ClientSecret = "2eb76a9352af6eca4001b20a5a84d86103467b1c9e5852bf95427b6b511695e6"
	RedirectURI = "http://localhost:1699/kick/callback"
	Scopes      = "user:read channel:read channel:write chat:read chat:write channel:read streamkey:read events:subscribe moderation:ban"
)

func HandleLogin() {
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
			http.Error(w, "Erro token: "+err.Error(), 500)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		helpers.Logf(helpers.Reset, "[KICK LOGIN] TOKEN: %s", body)
		var tokenResp struct {
			AccessToken  string `json:"access_token"`
			TokenType    string `json:"token_type"`
			RefreshToken string `json:"refresh_token"`
			ExpiresIn    int    `json:"expires_in"`
			Scope        string `json:"scope"`
		}
		json.Unmarshal(body, &tokenResp)

		helpers.Logf(helpers.Reset, "[KICK LOGIN] Login: access_token: %s; token_type: %s; refresh_token: %s; expires: %d; scope: %s", tokenResp.AccessToken, tokenResp.TokenType, tokenResp.RefreshToken, tokenResp.ExpiresIn, tokenResp.Scope)
		TokenMutex.Lock()
		Token = tokenResp.AccessToken
		RefreshToken = tokenResp.RefreshToken
		TokenMutex.Unlock()

		var userData, uErr = GetChannel(0, nil)
		if uErr != nil {
			log.Println("Erro ao obter info Kick:", uErr)
			http.Error(w, "Erro ao obter info do usuário", 500)
			return
		}

		UserID = userData.BroadcasterUserId
		UserLogin = userData.Slug

		close(LoginDone)
		fmt.Fprintf(w, "Login Kick concluído! Pode fechar esta página.\r\n")
		helpers.Logf(helpers.Reset, "[KICK LOGIN] Login concluído: %s (ID: %d)", UserLogin, UserID)

		if err := Connect(); err != nil {
			log.Fatal(err)
		}
	})
}

func GetKickToken() string {
	TokenMutex.RLock()
	defer TokenMutex.RUnlock()
	return Token
}
