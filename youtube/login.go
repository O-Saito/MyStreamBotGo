package youtube

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	Scopes = "https://www.googleapis.com/auth/youtube.readonly https://www.googleapis.com/auth/youtube.force-ssl"
)

type OAuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Error        string `json:"error"`
	ErrorDesc    string `json:"error_description"`
}

func HandleLogin() {

	if globals.GetConfig().YouTubeRefresh != "" {
		helpers.Logf("YOUTUBE!", "")
		user := globals.YouTubeUser{
			RefreshToken: globals.GetConfig().YouTubeRefresh,
		}

		globals.GetState().SetYouTubeUser(user)

		err := RefreshToken()
		if err == nil {
			yt, err := GetCurrentYouTubeChannel()
			if err == nil {
				for _, v := range yt.Items {
					globals.GetState().AddYouTubeChannel(globals.YouTubeChannel{
						ID:          v.ID,
						Title:       v.Snippet.Title,
						Thumbnail:   v.Snippet.Thumbnails.Medium.URL,
						Description: v.Snippet.Description,
						CustomURL:   v.Snippet.Description,
					})
				}

				newUser := globals.GetState().GetYouTubeUser()

				globals.WsBroadcast <- globals.SocketMessage{
					Type: "youtube-connection",
					Data: newUser,
				}
			} else {
				helpers.Logf(helpers.Red, "Falha ao buscar canais do YT %s", err.Error())
			}
		} else {
			helpers.Logf(helpers.Red, "Falha ao atualizar token %s", err.Error())
		}
	}

	redirectURI := fmt.Sprintf("http://localhost:%s/youtube/callback", globals.GetConfig().HTTPPort)

	scopes := Scopes

	/*if globals.GetConfig().TwitchScopes != "" {
		scopes = fmt.Sprintf("%s %s", scopes, globals.GetConfig().TwitchScopes)
	}
	subTypes := globals.GetConfig().GetTwitchSubTypes()
	for _, es := range subTypes {
		if es != nil && es["requires"] != nil {
			reqs := strings.SplitSeq(es["requires"].(string), " ")
			for req := range reqs {
				if strings.Contains(scopes, req) {
					continue
				}
				scopes = fmt.Sprintf("%s %s", scopes, req)
			}
		}
	}*/

	// Endpoint que redireciona para YT
	http.HandleFunc("/youtube/login", func(w http.ResponseWriter, r *http.Request) {
		authURL := fmt.Sprintf(
			"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&access_type=offline&prompt=consent",
			globals.GetConfig().YouTubeClientID,
			url.QueryEscape(redirectURI),
			url.QueryEscape(scopes),
		)
		helpers.Logf(helpers.Reset, "[YT LOGIN] Abrindo URL de login: %s", authURL)
		http.Redirect(w, r, authURL, http.StatusFound)
	})

	// Callback do Youtube
	http.HandleFunc("/youtube/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Code missing", 400)
			return
		}

		data := url.Values{}
		data.Set("client_id", globals.GetConfig().YouTubeClientID)
		data.Set("client_secret", globals.GetConfig().YouTubeClientSecret)
		data.Set("code", code)
		data.Set("grant_type", "authorization_code")
		data.Set("redirect_uri", redirectURI)

		resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
		if err != nil {
			http.Error(w, "Erro token: "+err.Error(), 500)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		var tokenResp OAuthResponse
		json.Unmarshal(body, &tokenResp)
		//Token := tokenResp.AccessToken

		user := globals.YouTubeUser{
			Token:          tokenResp.AccessToken,
			RefreshToken:   tokenResp.RefreshToken,
			TokenExpiresIn: tokenResp.ExpiresIn,
		}

		globals.GetState().SetYouTubeUser(user)

		helpers.Logf(helpers.Red, "[YT TOKEN] %s : %s E: %d", string(body), globals.GetConfig().YouTubeClientID, tokenResp.ExpiresIn)

		fmt.Fprintf(w, "Login concluído! Pode fechar esta página.")

		yt, err := GetCurrentYouTubeChannel()

		if err != nil {
			http.Error(w, "Erro buscar channel: "+err.Error(), 500)
			return
		}

		if len(yt.Items) == 0 {
			return
		}

		for _, v := range yt.Items {
			globals.GetState().AddYouTubeChannel(globals.YouTubeChannel{
				ID:          v.ID,
				Title:       v.Snippet.Title,
				Thumbnail:   v.Snippet.Thumbnails.Medium.URL,
				Description: v.Snippet.Description,
				CustomURL:   v.Snippet.Description,
			})
		}

		newUser := globals.GetState().GetYouTubeUser()

		globals.WsBroadcast <- globals.SocketMessage{
			Type: "youtube-connection",
			Data: newUser,
		}
	})
}
