package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const (
	//ClientID     = "jenisabhabc5zl01bhu86gcjeoe99z"
	//ClientSecret = "ddipe06ckokzhhc7fk5stu8ft8ybb9"
	RedirectURI = "http://localhost:1699/twitch/callback"
	Scopes      = "chat:read chat:edit user:read:email moderator:manage:chat_messages channel:moderate channel:read:subscriptions"
)

func HandleLogin() {

	scopes := Scopes

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
	}

	// Endpoint que redireciona para Twitch
	http.HandleFunc("/twitch/login", func(w http.ResponseWriter, r *http.Request) {
		authURL := fmt.Sprintf(
			"https://id.twitch.tv/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=%s",
			globals.GetConfig().TwitchClientID,
			url.QueryEscape(RedirectURI),
			url.QueryEscape(scopes),
		)
		helpers.Logf(helpers.Reset, "[TWITCH LOGIN] Abrindo URL de login: %s", authURL)
		http.Redirect(w, r, authURL, http.StatusFound)
	})

	// Callback da Twitch
	http.HandleFunc("/twitch/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Code missing", 400)
			return
		}

		data := url.Values{}
		data.Set("client_id", globals.GetConfig().TwitchClientID)
		data.Set("client_secret", globals.GetConfig().TwitchClientSecret)
		data.Set("code", code)
		data.Set("grant_type", "authorization_code")
		data.Set("redirect_uri", RedirectURI)

		resp, err := http.PostForm("https://id.twitch.tv/oauth2/token", data)
		if err != nil {
			http.Error(w, "Erro token: "+err.Error(), 500)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		var tokenResp struct {
			AccessToken string `json:"access_token"`
		}
		json.Unmarshal(body, &tokenResp)
		Token := tokenResp.AccessToken

		helpers.Logf("[Twitch TOKEN] %s : %s", Token, globals.GetConfig().TwitchClientID)

		// Pegar info do usuário
		req, _ := http.NewRequest("GET", "https://api.twitch.tv/helix/users", nil)
		req.Header.Set("Authorization", "Bearer "+Token)
		req.Header.Set("Client-ID", globals.GetConfig().TwitchClientID)
		userResp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, "Erro users: "+err.Error(), 500)
			return
		}
		defer userResp.Body.Close()
		dataUser, _ := io.ReadAll(userResp.Body)

		var u struct {
			Data []TwitchUserData `json:"data"`
		}
		json.Unmarshal(dataUser, &u)

		if len(u.Data) == 0 {
			helpers.Log(helpers.Red, "[TWITCH LOGIN] Erro: Nenhum usuário retornado. Verifique token e scopes.")
			return
		}

		d := u.Data[0]

		user := globals.TwitchUser{
			Token:                  Token,
			UserID:                 d.ID,
			UserLogin:              d.Login,
			DisplayName:            d.DisplayName,
			Type:                   d.Type,
			BroadcasterType:        d.BroadcasterType,
			Description:            d.Description,
			ProfileImageURL:        d.ProfileImageURL,
			ProfileOfflineImageURL: d.ProfileOfflineImageURL,
			ViewCount:              d.ViewCount,
			Email:                  d.Email,
			Connected:              true,
		}
		globals.GetState().SetTwitchUser(user)
		helpers.Logf(helpers.Reset, "[TWITCH LOGIN] UserID: %s, UserLogin: %s", user.UserID, user.UserLogin)

		fmt.Fprintf(w, "Login concluído! Pode fechar esta página.")

		if err := Connect(); err != nil {
			log.Fatal(err)
		}
		globals.WsBroadcast <- globals.SocketMessage{
			Type: "twitch-connection",
			Data: user,
		}
		JoinChannel(user.UserLogin)

		connectToEventSub()
	})
}
