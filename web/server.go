package web

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"MyStreamBot/kick"
	"MyStreamBot/twitch"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

// IrcHandler exportado para endpoints administrativos
//var IrcHandler *irc.ChatHandler

type SocketMessage struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var wsClients = make(map[*websocket.Conn]bool)

var SocketHandlers = map[string]func(map[string]any){}

func StartHTTPServer() {
	// Servir frontend
	http.Handle("/", http.FileServer(http.Dir("./web")))

	// WebSocket
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			helpers.Logf(helpers.Red, "[WebSocket] Erro: %s", err.Error())
			return
		}
		defer conn.Close()
		wsClients[conn] = true

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				delete(wsClients, conn)
				break
			}

			helpers.Logf(helpers.Cyan, "[Socket] Message: %s", string(msg))
			m := string(msg)
			if m == "init" {
				jsonData := map[string]any{
					"type": "init",
					"data": map[string]any{
						"twitch": map[string]any{
							"connected_as": twitch.UserLogin,
						},
						"kick": map[string]any{
							"connected_as": kick.UserLogin,
						},
						"twitch_connected_chat": twitch.Channels,
						"kick_connected_chat":   kick.Channels,
					},
				}
				helpers.Log(helpers.Cyan, "[Socket] Init message")
				d, err := json.Marshal(jsonData)
				if err != nil {
					helpers.Logf(helpers.Red, "[Socket] Init error: %s", err.Error())
					continue
				}
				helpers.Logf(helpers.Cyan, "[Socket] Init message: %s", d)
				conn.WriteMessage(websocket.TextMessage, []byte(d))
				continue
			}
			var data SocketMessage
			json.Unmarshal([]byte(m), &data)
			if handler, exists := SocketHandlers[string(data.Type)]; exists {
				handler(data.Data)
				continue
			}
		}
	})

	// Endpoints administrativos
	http.HandleFunc("/admin/delete/twitch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Método inválido", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Message string `json:"message"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		twitch.DeleteMessage(req.Message)
		//if IrcHandler != nil {
		//	IrcHandler.SendMessage("/delete " + req.Message)
		//}
		w.WriteHeader(200)
	})

	http.HandleFunc("/admin/ban/twitch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Método inválido", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			User string `json:"user"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		//if IrcHandler != nil {
		//	IrcHandler.SendMessage("/ban " + req.User)
		//}
		w.WriteHeader(200)
	})

	// Goroutine para enviar mensagens do backend para todos os clients
	go func() {
		for msg := range globals.WsBroadcast {
			helpers.Logf(helpers.Cyan, "[WebSocket] Broadcast: %s - %s", msg.Type, msg.Data)
			for client := range wsClients {
				message := fmt.Sprintf(`{"type":"%s","data":%s}`, msg.Type, msg.Data)
				client.WriteMessage(websocket.TextMessage, []byte(message))
			}
		}
	}()

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}

	for _, addr := range addrs {
		// Verifica se é IP do tipo IPNet e não loopback
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			// Só IPv4
			if ipnet.IP.To4() != nil {
				helpers.Logf(helpers.Reset, "IP local da máquina: %s", ipnet.IP.String())
			}
		}
	}
	log.Println("[MyStreamBot] Servidor HTTP iniciado em http://localhost:1699")
	go http.ListenAndServe("0.0.0.0:1699", nil)
}
