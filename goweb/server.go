package goweb

import (
	"MyStreamBot/globals"
	"MyStreamBot/helpers"
	"MyStreamBot/kick"
	"MyStreamBot/mlua"
	"MyStreamBot/twitch"
	"encoding/json"
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
var wsClientsUpgraded = make(map[string][]*websocket.Conn)

var SocketHandlers = map[string]func(*websocket.Conn, map[string]any){
	"init": func(c *websocket.Conn, m map[string]any) {
		jsonData := SocketMessage{
			Type: "init",
			Data: map[string]any{
				"twitch": globals.GetState().GetTwitchUser(),
				"kick": map[string]any{
					"connected_as": kick.UserLogin,
				},
				"twitch_connected_chat": twitch.Channels,
				"kick_connected_chat":   kick.Channels,
				"custom_events_modules": mlua.ListDynamicEvents(),
				"twitch_eventsubs":      globals.GetState().GetData("TwitchSubEventsConnectedEvents"),
			},
		}
		helpers.Log(helpers.Cyan, "[Socket] Init message")
		d, err := json.Marshal(jsonData)
		if err != nil {
			helpers.Logf(helpers.Red, "[Socket] Init error: %s", err.Error())
			return
		}
		helpers.Logf(helpers.Cyan, "[Socket] Init message: %s", d)
		c.WriteMessage(websocket.TextMessage, []byte(d))
	},
	"upgrade-conn": func(c *websocket.Conn, m map[string]any) {
		data := map[string]any{
			"type": "response-upgrade",
			"data": "",
		}
		if m["conn"] != nil {
			if wsClientsUpgraded[m["conn"].(string)] == nil {
				wsClientsUpgraded[m["conn"].(string)] = make([]*websocket.Conn, 0)
			}

			wsClientsUpgraded[m["conn"].(string)] = append(wsClientsUpgraded[m["conn"].(string)], c)
			data["data"] = "conexão atualizada!"
		} else {
			data["data"] = "conexão não especificada!"
		}

		d, err := json.Marshal(data)
		if err != nil {
			helpers.Logf(helpers.Red, "[Socket] Upgrade-conn parse error: %s", err.Error())
			return
		}
		c.WriteMessage(websocket.TextMessage, []byte(d))
	},
}

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
				SocketHandlers["init"](conn, nil)
				continue
			}
			var data SocketMessage
			json.Unmarshal([]byte(m), &data)
			if handler, exists := SocketHandlers[string(data.Type)]; exists {
				handler(conn, data.Data)
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

			if msg.Filter != "" {
				helpers.Logf(helpers.Cyan, "[WebSocket] Message filter %s: %s - %s", msg.Filter, msg.Type, msg.Data)
				wsList := wsClientsUpgraded[msg.Filter]
				if len(wsList) == 0 {
					continue
				}
				jsonData, _ := json.Marshal(msg)
				for _, client := range wsList {
					client.WriteMessage(websocket.TextMessage, []byte(jsonData))
				}
				continue
			}

			helpers.Logf(helpers.Cyan, "[WebSocket] Broadcast: %s - %s", msg.Type, msg.Data)
			jsonData, _ := json.Marshal(msg)
			for client := range wsClients {
				client.WriteMessage(websocket.TextMessage, []byte(jsonData))
			}
		}
	}()

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}

	log.Printf("[MyStreamBot] Possíveis IP's (para os logins TEM que ser pelo localhost):")
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				helpers.Logf(helpers.Reset, "http://%s:1699", ipnet.IP.String())
			}
		}
	}
	log.Println("[MyStreamBot] Servidor HTTP iniciado em http://localhost:1699")
	go http.ListenAndServe("0.0.0.0:1699", nil)
}
