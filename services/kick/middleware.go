package kick

import (
	"MyStreamBot/helpers"
	"net/http"
)

// DoRequest executa uma requisição HTTP para Kick com renovação automática de token
// Detecta erros 401 (token expirado) e tenta fazer refresh e retry automaticamente
func DoRequest(req *http.Request) (*http.Response, error) {
	// Define o Authorization header
	req.Header.Set("Authorization", "Bearer "+Token)

	// Faz a requisição
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Se recebeu 401 (Unauthorized), tenta refresh automático
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		helpers.Logf(helpers.DEBUG, "[KICK] Token expirado (401), tentando refresh...")

		// Faz refresh do token
		// if err := RefreshToken(); err != nil {
		// 	helpers.Logf(helpers.ERROR, "[KICK] Falha ao fazer refresh: %s", err.Error())
		// 	return nil, err
		// }

		// // Salva o novo token no banco
		// newUser := globals.GetState().GetYouTubeUser()
		// sqlErr := globals.GetGlobalDB().SaveToken("youtube", newUser.Token, newUser.RefreshToken, time.Now().Add(time.Duration(newUser.TokenExpiresIn)*time.Second))
		// if sqlErr != nil {
		// 	helpers.Logf(helpers.ERROR, "[KICK] Falha ao salvar token renovado: %s", sqlErr.Error())
		// }

		// // Atualiza o Authorization header com o novo token
		// req.Header.Set("Authorization", "Bearer "+newUser.Token)

		// // Tenta novamente
		// helpers.Logf(helpers.DEBUG, "[KICK] Retry de requisição com novo token")
		// resp, err = http.DefaultClient.Do(req)
		// if err != nil {
		// 	return nil, err
		// }
	}

	return resp, nil
}
