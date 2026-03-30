package twitch

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"net/http"
)

// DoRequest executa uma requisição HTTP para Twitch com renovação automática de token
// Detecta erros 401 (token expirado) e tenta fazer refresh e retry automaticamente
func DoRequest(req *http.Request) (*http.Response, error) {
	user := globals.GetState().GetTwitchUser()

	// Define o Authorization header
	req.Header.Set("Authorization", "Bearer "+user.Token)

	// Faz a requisição
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Se recebeu 401 (Unauthorized), tenta refresh automático
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		helpers.Logf(helpers.DEBUG, "[TWITCH] Token expirado (401), tentando refresh...")

		currentAccess, err := globals.GetGlobalDB().GetToken("twitch")
		_, err = RefreshToken(currentAccess.RefreshToken)
		// Faz refresh do token
		if err != nil {
			helpers.Logf(helpers.ERROR, "[TWITCH] Falha ao fazer refresh: %s", err.Error())
			return nil, err
		}

		newUser := globals.GetState().GetTwitchUser()

		// Atualiza o Authorization header com o novo token
		req.Header.Set("Authorization", "Bearer "+newUser.Token)

		// Tenta novamente
		helpers.Logf(helpers.DEBUG, "[TWITCH] Retry de requisição com novo token")
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}
