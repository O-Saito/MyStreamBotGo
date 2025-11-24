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
	"slices"
	"strings"
	"time"
)

const (
	Scopes = "chat:read chat:edit user:read:email moderator:manage:chat_messages moderator:manage:banned_users channel:moderate channel:read:subscriptions"
)

func HandleLogin() {

	redirectURI := fmt.Sprintf("http://localhost:%s/twitch/callback", globals.GetConfig().HTTPPort)

	scopes := Scopes

	if globals.GetConfig().TwitchScopes != "" {
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
	}

	currentAccess, err := globals.GetGlobalDB().GetToken("twitch")
	if err != nil {
		helpers.Logf(helpers.Red, "[TWITCH HANDLELOGIN] Falha ao buscar access token %s", err.Error())
	}

	if currentAccess != nil {
		data, err := ValidateAccessToken(currentAccess.AccessToken)
		helpers.Logf(helpers.Red, "[TWITCH LOG VALIDATE ACCESS TOKEN] %v", data)
		if err != nil {
			helpers.Logf(helpers.Red, "[TWITCH HANDLELOGIN] Falha ao validar access token %s", err.Error())
		}

		reconnect := true
		if data != nil && data.Status != 401 { // status 401 = invalid
			scopesValid := true
			//missingScopes := []string{}
			for v := range strings.SplitSeq(scopes, " ") {
				if !slices.Contains(data.Scopes, v) {
					scopesValid = false
					//missingScopes = append(missingScopes, v)
					break
				}
			}

			if scopesValid && data.ExpiresIn > 0 {
				reconnect = false
				err := initTwitchUser(currentAccess.AccessToken)
				if err != nil {
					helpers.Logf(helpers.Red, "[TWITCH HANDLELOGIN] Falha ao inicializar twitch user")
				}
			}
		}

		if reconnect {
			d, err := refreshToken(currentAccess.RefreshToken)
			if err != nil {
				helpers.Logf(helpers.Red, "[TWITCH HANDLELOGIN] Falha ao buscar token %s", err.Error())
			}
			if d != nil {
				sqlErr := globals.GetGlobalDB().SaveToken("twitch", d.AccessToken, d.RefreshToken, time.Now().Add(time.Duration(d.Expires)*time.Second))
				if sqlErr != nil {
					helpers.Logf(helpers.Red, "[TWITCH HANDLELOGIN] Falha ao salvar token %s", sqlErr.Error())
				}
				initTwitchUser(d.AccessToken)
			}
		}
	}

	// Endpoint que redireciona para Twitch
	http.HandleFunc("/twitch/login", func(w http.ResponseWriter, r *http.Request) {
		authURL := fmt.Sprintf(
			"https://id.twitch.tv/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=%s",
			globals.GetConfig().TwitchClientID,
			url.QueryEscape(redirectURI),
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
		data.Set("redirect_uri", redirectURI)

		resp, err := http.PostForm("https://id.twitch.tv/oauth2/token", data)
		if err != nil {
			http.Error(w, "Erro token: "+err.Error(), 500)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		var tokenResp struct {
			AccessToken  string `json:"access_token"`
			Expires      int    `json:"expires_in"`
			RefreshToken string `json:"refresh_token"`
		}

		json.Unmarshal(body, &tokenResp)
		Token := tokenResp.AccessToken

		sqlErr := globals.GetGlobalDB().SaveToken("twitch", tokenResp.AccessToken, tokenResp.RefreshToken, time.Now().Add(time.Duration(tokenResp.Expires)*time.Second))

		if sqlErr != nil {
			helpers.Logf(helpers.Red, "[TWITCH LOGIN] Falha ao salvar refresh token %s", sqlErr.Error())
		}

		initError := initTwitchUser(Token)
		if initError != nil {
			http.Error(w, "Falha ao fazer login: "+initError.Error(), http.StatusTeapot)
			return
		}

		fmt.Fprintf(w, "Login concluído! Pode fechar esta página.")
	})
}

func refreshToken(refreshToken string) (*struct {
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
	body, _ := io.ReadAll(resp.Body)

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		Expires      int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
	}

	json.Unmarshal(body, &tokenResp)
	return &tokenResp, nil
}

func initTwitchUser(token string) error {
	globals.GetState().SetTwitchUser(globals.TwitchUser{
		Token: token,
	})
	d, err := GetStreamerData()
	if err != nil {
		return err
	}

	user := globals.TwitchUser{
		Token:                  token,
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
	streamData, _ := GetChannelStreamData(user.UserID)
	if streamData != nil {
		user.StreamDetails = &globals.TwitchStreamData{
			GameId:      streamData.GameID,
			GameName:    streamData.GameName,
			Title:       streamData.Title,
			Tags:        streamData.Tags,
			ViewerCount: 0,
			Language:    streamData.BroadcasterLanguage,
		}
	}
	globals.GetState().SetTwitchUser(user)

	if err := Connect(); err != nil {
		log.Fatal(err)
	}
	globals.WsBroadcast <- globals.SocketMessage{
		Type: "twitch-connection",
		Data: user,
	}
	JoinChannel(user.UserLogin)
	connectToEventSub()
	return nil
}
